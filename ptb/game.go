package ptb

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Game tweaking
const (
	tweak_JOIN_DURATION = 30   // Time to wait for joins in seconds.
	tweak_MIN_DURATION  = 30   // Minimum game duration in seconds.
	tweak_MAX_DURATION  = 600  // Maximum game duration in seconds.
	tweak_DEFUSE        = true // Wether to enable bomb defusing.
	tweak_DEFUSE_CHANCE = 90   // Chance that a bomb can be defused: 0=never; 99=always.
	tweak_MIN_WIRES     = 4    // Minimum number of wires in the defuse minigame.
	tweak_MAX_WIRES     = 12   // Maximum number of wires in the defuse minigame.
	tweak_FAKE          = true // Enable fake bombs.
	tweak_FAKE_CHANCE   = 10   // Chance that a bomb will be fake: 0=never; 99=always.
	tweak_MIN_PLAYERS   = 4    // Minimum number of players.

	tweak_KICK     = true // Kick player on explosion.
	tweak_BAN      = true // Ban player after explosion. (prevent auto rejoin)
	tweak_BAN_TIME = 10   // Ban time in seconds.
)

// Game states
const (
	state_INIT    = iota // Nothing?
	state_WARMUP         // Players can join, game is being explained.
	state_PLAYING        // Bomb is being passed around.
	state_ENDED          // Game has ended.
)

// Defuse states
const (
	defuse_NOTHING   = iota // Nothing happens
	defuse_LESS_TIME        // Timer goes down
	defuse_MORE_TIME        // Timer goes up!
	defuse_EXPLODE          // Bomb explodes
	defuse_SUCCESS          // Bomb succesfully defused!
	defuse_CUT              // Wire is already cut.
)

// Chat provides methods to communicate with players.
type Chat interface {
	Public(message string)        // Sends a message to all players.
	Private(nick, message string) // Sends a message to given player.
	Kick(nick, reason string)     // Kicks a player.
	IsOperator() bool             // True if the bot has the power to kick people.
	Ban(nick string) bool         // Ban given nickname.
	UnBan(nick string)            // Unban given nickname.
}

// bomb represents a bomb being thrown around
type bomb struct {
	fake       bool      // Fake bombs do not actually explode
	defusable  bool      // True if this bomb can be defused
	wires      []uint8   // Wire functions
	location   *Player   // Player currently holding the bomb
	detonation time.Time // Detonation time
	throwTime  time.Time // Last time the bomb was thrown
	defused    bool
}

// randomize sets a random detonation time.
func (b *bomb) randomize(min, max time.Duration) {
	duration := time.Duration((rand.Int63n(int64(max)) + int64(min)))
	b.detonation = time.Now().Add(duration)
}

// turn represents a time the player had a bomb.
// We keep track of actions a certain player does to gather
// statistics usefull when calculating scores.
type Turn struct {
	source        *Player       // Who threw the bomb to this player?
	target        *Player       // Where did the bomb go after this turn?
	Duration      time.Duration // How long did the player keep the bomb?
	Time          time.Time     // When did this turn happen?
	DefuseAttempt bool          // Did the player defuse during this turn?

	SourceNick string // Nickname of the source for JSON export.
	TargetNick string // Nickname of the target for JSON export.
}

// player represents a single player in the game
type Player struct {
	Nick          string  // Actual nickname of this player
	sanitizedNick string  // Sanitized nickname
	Late          bool    // True if this player joined after the game started
	turns         []*Turn // List of turns this player has played.
	DefuseAttempt bool    // True if player already tried to defuse.
	Defused       bool    // Player defused!
	Dead          bool    // Bomb exploded while the player was holding it.

	Duration     time.Duration // Total turn duration for JSON export.
	MeanDuration time.Duration // Mean turn duration for JSON export.
	Turns        uint          // Number of turns for JSON export.
}

// Game represents a single instance of the game.
type Game struct {
	bomb    *bomb              // The bomb used in this game
	Players map[string]*Player // Players
	state   uint8              // Game state
	chat    Chat               // Interface to the chatroom
	first   *Player            // First player to start
	stop    chan bool          // Indicates the game ended.
	turn    *Turn              // Current turn, or nil if the bomb was dropped.
	Started time.Time          // Game start time.
	Ended   time.Time          // Game end time.

	Turns []*Turn // Complete list of turns for JSON export.
}

func NewGame(chat Chat) *Game {
	g := new(Game)
	g.chat = chat

	return g
}

func (g *Game) nextTurn(next *Player) {

	// Finalize last turn.
	if g.turn != nil {
		g.turn.target = next
		g.turn.TargetNick = next.Nick

		// Calculate time
		g.turn.Duration = time.Now().Sub(g.turn.Time)
		g.turn.source.Duration = g.turn.source.Duration + g.turn.Duration
	}

	// Bomp dropped, no next turn.
	if next == nil {
		g.turn = nil
		g.bomb.location = nil
		return
	}

	last := g.turn.source

	// Initialize new turn
	g.turn = new(Turn)
	g.turn.source = last
	g.turn.SourceNick = last.Nick
	g.turn.Time = time.Now()

	// Add pointer to the player's turn list.
	next.turns = append(next.turns, g.turn)
	g.Turns = append(g.Turns, g.turn)

	// Update bomb location.
	// TODO: Delete this var and always read location from current turn.
	g.bomb.location = next

	return

}

// sanitizeNick returns a lowercase version of the nickname, stripped from spaces.
func sanitizeNick(nick string) string {
	return strings.ToLower(strings.Trim(nick, " \t\n\r"))
}

// IsActive returns true if there is possible interaction with players.
func (g *Game) IsActive() bool {
	return (g.state != state_INIT && g.state != state_ENDED)
}

// Start launches the joining timeslot for a new game!
func (g *Game) Start() {

	if g.IsActive() || g.chat == nil {
		return
	}

	g.bomb = new(bomb)

	// Choose fake bombs
	if tweak_FAKE {
		g.bomb.fake = (rand.Intn(100) <= tweak_FAKE_CHANCE)
	}

	// Prepare defuse minigame
	if tweak_DEFUSE {
		g.bomb.defusable = (rand.Intn(100) <= tweak_DEFUSE_CHANCE)

		n := rand.Intn(tweak_MAX_WIRES-tweak_MIN_WIRES) + tweak_MIN_WIRES
		g.bomb.wires = make([]uint8, n)

		for i := range g.bomb.wires {
			g.bomb.wires[i] = uint8(rand.Intn(5))
		}
	}

	// Make sure we reset everything before starting a new game.
	g.Started = time.Now()
	g.bomb.detonation = time.Now()
	g.bomb.throwTime = time.Now()
	g.Players = make(map[string]*Player)
	g.Turns = make([]*Turn, 0, 10)
	g.state = state_WARMUP

	g.chat.Public(text_START_ATTENTION)
	g.chat.Public(text_START_JOIN)

	duration := (tweak_JOIN_DURATION / 5 * time.Second)

	// Explain how the game works during join time.
	for i := time.Duration(1); i < 5; i++ {
		g.explain(i, duration*i)
	}

	timer := time.NewTimer(duration * 5)

	// Wait for timer to expire
	<-timer.C

	g.start()

}

func (g *Game) explain(tick, duration time.Duration) {

	timer := time.NewTimer(duration)

	// Wait for timer to expire
	<-timer.C

	switch tick {

	case 1:
		g.chat.Public(text_HELP_THROW)
	case 2:
		g.chat.Public(text_HELP_SCORE)
	case 3:
		g.chat.Public(text_HELP_DEFUSE)
	case 4:
		if len(g.Players) >= tweak_MIN_PLAYERS {
			g.chat.Public(fmt.Sprintf(text_HELP_START, g.first.Nick))
		}

	}
}

// start is an internal method and starts the actual game after the joining timeslot.
func (g *Game) start() {

	if len(g.Players) < tweak_MIN_PLAYERS {
		g.chat.Public(text_START_FAIL)
		g.state = state_INIT
		return
	}

	// Send the bomb to the next player!
	g.nextTurn(g.first)
	g.bomb.location = g.first

	g.state = state_PLAYING

	g.bomb.randomize(tweak_MIN_DURATION*time.Second, tweak_MAX_DURATION*time.Second)

	// Send message.
	g.chat.Public(fmt.Sprintf(text_START_GO, g.first.Nick))

	ticker := time.NewTicker(10 * time.Second)

	// Start game ticks
	go func() {

		for {
			select {

			case <-ticker.C:

				// Check if the bomb has to explode.
				if time.Now().After(g.bomb.detonation) {
					g.Stop()
				}

			case <-g.stop:
				return

			}
		}

	}()

}

// Join adds a player to the game.
// Players can only join during the warmup or playing states.
func (g *Game) Join(nick string) {

	s := sanitizeNick(nick)

	// Fast path, joining is not allowed or useless.
	if (g.state != state_WARMUP && g.state != state_PLAYING) || g.Players[s] != nil {
		return
	}

	// Initialize new player
	p := new(Player)
	p.Nick = nick
	p.sanitizedNick = s
	p.Late = (g.state == state_PLAYING)
	p.turns = make([]*Turn, 0, 5)

	// Append to player map
	g.Players[s] = p

	if len(g.Players) <= 1 {
		g.first = p
	}

	// Notify everyone if this player joined after the game started
	if p.Late {
		g.chat.Public(fmt.Sprintf(text_PLAYER_JOINED_LATE, nick))
	} else {
		g.chat.Private(nick, text_PLAYER_JOINED)
	}

}

// Throw sends the bomb to another player.
// Only the player currently holding the bomb can throw.
func (g *Game) Throw(source, target string) {

	source = sanitizeNick(source)
	p := g.bomb.location

	// Fast path
	if g.state != state_PLAYING || p == nil || p.sanitizedNick != source {
		return
	}

	target = sanitizeNick(target)

	if source == target {
		g.chat.Public(fmt.Sprintf(text_BOMB_THROWN_SELF, p.Nick))
		return
	}

	// Attempt to fetch target
	t, playing := g.Players[target]

	// We can't throw to someone who doesn't play.
	if !playing {
		g.chat.Public(fmt.Sprintf(text_BOMB_DROPPED, target))
		g.nextTurn(nil)
		return
	}

	g.nextTurn(t)

	// Send message.
	g.chat.Public(fmt.Sprintf(text_BOMB_THROWN, p.Nick, t.Nick))

	return
}

// Pickup the bomb if it was on the ground.
func (g *Game) Pickup(nick string) {

	// Fast path
	if g.state != state_PLAYING || g.bomb.location != nil {
		return
	}

	nick = sanitizeNick(nick)

	// Attempt to fetch target
	t, playing := g.Players[nick]

	// Someone who isn't playing can't pick up the bomb.
	if !playing {
		return
	}

	g.nextTurn(t)

	// Send message.
	g.chat.Public(fmt.Sprintf(text_BOMB_PICKED_UP, t.Nick))

	return
}

// Defuse starts a defuse attempt.
// Only the player holding the bomb can try to defuse it.
// Once defusing started, a player has to cut a wire, the bomb can no
// longer be thrown around.
func (g *Game) Defuse(nick string) {

	nick = sanitizeNick(nick)
	p := g.bomb.location

	// Fast path
	if g.state != state_PLAYING || p == nil || p.sanitizedNick != nick {
		return
	}

	// TODO: Duplicate code in Defuse and Cut.

	if p.DefuseAttempt {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_TRIED, p.Nick))
		g.state = state_PLAYING
		return
	}

	// Check if the bomb can be defused!
	if !g.bomb.defusable {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_DISABLED, p.Nick))
		return
	}

	// Show defuse info message
	g.chat.Public(fmt.Sprintf(text_DEFUSE, len(g.bomb.wires)))

}

// Cut tries to cut a wire during defuse.
func (g *Game) Cut(nick string, wire uint8) {

	nick = sanitizeNick(nick)
	p := g.bomb.location

	// Fast path
	if (g.state != state_PLAYING) || p == nil || p.sanitizedNick != nick || !g.bomb.defusable {
		return
	}

	// Check if the wire exists
	if wire < 1 || wire > uint8(len(g.bomb.wires)) {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_ERROR, wire))
		g.state = state_PLAYING
		return
	}
	wire = wire - 1

	if p.DefuseAttempt {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_TRIED, p.Nick))
		g.state = state_PLAYING
		return
	}

	// Mark this player
	p.DefuseAttempt = true
	g.turn.DefuseAttempt = true

	// Check the wire function
	switch g.bomb.wires[wire] {

	case defuse_SUCCESS:
		g.bomb.defused = true
		g.chat.Public(fmt.Sprintf(text_DEFUSE_SUCCESS, p.Nick))
		g.Stop()
		return

	case defuse_NOTHING:
		g.chat.Public(fmt.Sprintf(text_DEFUSE_NOTHING, p.Nick))

	case defuse_LESS_TIME:
		g.bomb.randomize(20*time.Second, 60*time.Second)
		g.chat.Public(text_DEFUSE_LESS_TIME)

	case defuse_MORE_TIME:
		d := g.bomb.detonation.Sub(time.Now())
		g.bomb.randomize(d, d+(5*time.Minute))
		g.chat.Public(text_DEFUSE_MORE_TIME)

	case defuse_EXPLODE:
		g.Stop()
		return

	case defuse_CUT:
		g.chat.Public(fmt.Sprintf(text_DEFUSE_DUPLICATE, wire))

	}

	g.bomb.wires[wire] = defuse_CUT

	return

}

// Stop ends the game, bomb explodes or turns out to be fake.
func (g *Game) Stop() {

	if g.state != state_PLAYING {
		return
	}

	// The game ended!
	g.Ended = time.Now()
	g.state = state_ENDED

	if !g.bomb.defused {

		g.bomb.location.Dead = true

		// Check if the bomb was lying on the ground at detonation time.
		if g.bomb.location == nil {

			g.chat.Public(text_BOMB_EXPLODE)

		} else {

			// Show message and kick players if we can.
			if g.bomb.fake {
				g.chat.Public(text_BOMB_FAKE)
			} else if !tweak_KICK || !g.chat.IsOperator() {
				g.chat.Public(text_BOMB_EXPLODE)
				g.chat.Public(fmt.Sprintf(text_BOMB_EXPLODE_NOOP, g.bomb.location.Nick))
			} else {

				if tweak_BAN && g.chat.Ban(g.bomb.location.Nick) {

					// Schedule unban!
					go func() {
						timer := time.NewTimer(tweak_BAN_TIME * time.Second)

						// Wait for timer to expire
						<-timer.C

						g.chat.UnBan(g.bomb.location.Nick)
					}()
				}

				g.chat.Kick(g.bomb.location.Nick, text_BOMB_EXPLODE)
			}

		}

	}

	g.stop <- true

}

// Leave removes a player after he has left the room.
func (g *Game) Leave(nick string) {

	if g.state == state_INIT || g.state == state_ENDED {
		return
	}

	nick = sanitizeNick(nick)

	// Attempt to fetch the player
	p, playing := g.Players[nick]

	// This one isn't playing, couldn't care less.
	if !playing {
		return
	}

	g.chat.Public(fmt.Sprintf(text_PLAYER_LEFT, p.Nick))

	delete(g.Players, nick)

}

// Rename changes the name of a player.
func (g *Game) Rename(old, nick string) {

	if g.state == state_INIT || g.state == state_ENDED {
		return
	}

	olds := sanitizeNick(old)

	// Attempt to fetch the player
	p, playing := g.Players[olds]

	// This one isn't playing, couldn't care less.
	if !playing {
		return
	}

	p.Nick = nick
	p.sanitizedNick = sanitizeNick(nick)

	delete(g.Players, olds)
	g.Players[p.sanitizedNick] = p

	g.chat.Public(fmt.Sprintf(text_PLAYER_RENAME, old, p.Nick))

}

// Players shows a list of current players in the channel.
func (g *Game) PlayerList() {

	if g.state != state_PLAYING {
		return
	}

	players := make([]string, len(g.Players))

	i := 0
	for _, player := range g.Players {
		players[i] = player.Nick
		i++
	}

	g.chat.Public(fmt.Sprintf(text_PLAYER_LIST, strings.Join(players, ", ")))

}

// Decode attempts to find game commands in given message.
// This provides access to everything except starting the game.
func (g *Game) Decode(sender, message string) {

	if len(message) < 3 || !strings.HasPrefix(message, cmd_PREFIX) {
		return
	}

	args := strings.Split(message[1:], " ")

	switch args[0] {

	case cmd_JOIN:
		g.Join(sender)

	case cmd_PASS:
		if len(args) > 1 {
			g.Throw(sender, args[1])
		}

	case cmd_DEFUSE:
		g.Defuse(sender)

	case cmd_CUT:
		if len(args) > 1 {
			if n, err := strconv.ParseUint(args[1], 10, 8); err == nil {
				g.Cut(sender, uint8(n))
			}
		}

	case cmd_PLAYER_LIST:
		g.PlayerList()

	case cmd_PICK_UP:
		g.Pickup(sender)

	}

}
