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

// TODO: Publish events using go channels
// TODO: Live statuspage over HTTP (websockets?)
// TODO: Provide a JSON log file with detailed scores and statistics

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
	state_DEFUSE         // A player is trying to defuse the bomb.
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
	Op() bool                     // True if the bot has the power to kick people.
	Ban(nick string) bool         // Ban given nickname.
	UnBan(nick string)            // Unban given nickname.
}

// bomb represents a bomb being thrown around
type bomb struct {
	fake       bool      // Fake bombs do not actually explode
	defusable  bool      // True if this bomb can be defused
	wires      []uint8   // Wire functions
	location   *player   // Player currently holding the bomb
	detonation time.Time // Detonation time
	throwTime  time.Time // Last time the bomb was thrown
	defused bool
}

// time returns the duration since last time it was thrown.
// The timer is reset to prepare for a next throw.
func (b *bomb) time() time.Duration {
	now := time.Now()
	duration := now.Sub(b.throwTime)
	b.throwTime = now

	return duration
}

// randomize sets a random detonation time.
func (b *bomb) randomize(min, max time.Duration) {
	duration := time.Duration((rand.Int63n(int64(max)) + int64(min)))
	b.detonation = time.Now().Add(duration)
}

// turn represents a time the player had a bomb.
// We keep track of actions a certain player does to gather
// statistics usefull when calculating scores.
type turn struct {
	source   *player       // Who threw the bomb to this player?
	target   *player       // Where did the bomb go after this turn?
	duration time.Duration // How long did the player keep the bomb?
	time     time.Time     // When did this turn happen?
	defused  bool          // Did the player defuse during this turn?
}

// player represents a single player in the game
type player struct {
	nick          string // Actual nickname of this player
	sanitizedNick string // Sanitized nickname
	late          bool   // True if this player joined after the game started
	turns         []*turn
	current       *turn
	defuseAttempt bool // True if player already tried to defuse.
	defused       bool // Player defused!
	dead bool
}

// catch indicates that this player has catched the bomb from another player
func (p *player) catch(source *player) *turn {
	source.current.target = p

	t := new(turn)
	t.source = source
	t.time = time.Now()

	// Add pointers to the player object
	p.turns = append(p.turns, t)
	p.current = t

	return t
}

// Game represents a single instance of the game.
type Game struct {
	bomb    *bomb              // The bomb used in this game
	players map[string]*player // Players
	state   uint8              // Game state
	chat    Chat               // Interface to the chatbox
	first   *player            // First player to start
	stop    chan bool          // Indicates the game ended.
	started time.Time          // Game start time.
	ended   time.Time          // Game end time.
}

func (g *Game) SetChat(chat Chat) {
	g.chat = chat
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

	if (g.state != state_INIT && g.state != state_ENDED) || g.chat == nil {
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
	g.started = time.Now()
	g.bomb.detonation = time.Now()
	g.bomb.throwTime = time.Now()
	g.players = make(map[string]*player)
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
		if len(g.players) >= tweak_MIN_PLAYERS {
			g.chat.Public(fmt.Sprintf(text_HELP_START, g.first.nick))
		}

	}
}

// start is an internal method and starts the actual game after the joining timeslot.
func (g *Game) start() {

	if len(g.players) < tweak_MIN_PLAYERS {
		g.chat.Public(text_START_FAIL)
		g.state = state_INIT
		return
	}

	// Send the bomb to the next player!
	g.first.catch(nil)
	g.bomb.location = g.first

	g.state = state_PLAYING

	g.bomb.randomize(tweak_MIN_DURATION*time.Second, tweak_MAX_DURATION*time.Second)

	// Send message.
	g.chat.Public(fmt.Sprintf(text_START_GO, g.first.nick))

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
	if (g.state != state_WARMUP && g.state != state_PLAYING) || g.players[s] != nil {
		return
	}

	// Initialize new player
	p := new(player)
	p.nick = nick
	p.sanitizedNick = s
	p.late = (g.state == state_PLAYING)
	p.turns = make([]*turn, 0, 5)

	// Append to player map
	g.players[s] = p

	if len(g.players) <= 1 {
		g.first = p
	}

	// Notify everyone if this player joined after the game started
	if p.late {
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
		g.chat.Public(fmt.Sprintf(text_BOMB_THROWN_SELF, p.nick))
		return
	}

	// Attempt to fetch target
	t, playing := g.players[target]

	// Calculate how long the current player had the bomb.
	p.current.duration = g.bomb.time()

	// We can't throw to someone who doesn't play.
	if !playing {
		g.chat.Public(fmt.Sprintf(text_BOMB_DROPPED, target))

		// Nobody has the bomb
		g.bomb.location = nil

		return
	}

	// Send the bomb to the next player!
	t.catch(p)
	g.bomb.location = t

	// Send message.
	g.chat.Public(fmt.Sprintf(text_BOMB_THROWN, p.nick, t.nick))

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
	t, playing := g.players[nick]

	// Someone who isn't playing can't pick up the bomb.
	if !playing {
		return
	}

	// Send the bomb to the next player!
	t.catch(nil)
	g.bomb.location = t

	// Send message.
	g.chat.Public(fmt.Sprintf(text_BOMB_PICKED_UP, t.nick))

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

	if p.defuseAttempt {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_TRIED, p.nick))
		g.state = state_PLAYING
		return
	}

	// Check if the bomb can be defused!
	if !g.bomb.defusable {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_DISABLED, p.nick))
		return
	}

	// Show defuse info message
	g.chat.Public(fmt.Sprintf(text_DEFUSE, len(g.bomb.wires)))
	g.state = state_DEFUSE

}

// Cut tries to cut a wire during defuse.
func (g *Game) Cut(nick string, wire uint8) {

	nick = sanitizeNick(nick)
	p := g.bomb.location

	// Fast path
	if (g.state != state_DEFUSE && g.state != state_PLAYING) || p == nil || p.sanitizedNick != nick || !g.bomb.defusable {
		return
	}

	// Check if the wire exists
	if wire < 1 || wire > uint8(len(g.bomb.wires)) {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_ERROR, wire))
		g.state = state_PLAYING
		return
	}
	wire = wire - 1

	if p.defuseAttempt {
		g.chat.Public(fmt.Sprintf(text_DEFUSE_TRIED, p.nick))
		g.state = state_PLAYING
		return
	}

	// Mark this player
	p.defuseAttempt = true
	p.current.defused = true

	// Check the wire function
	switch g.bomb.wires[wire] {

	case defuse_SUCCESS:
		g.bomb.defused = true
		g.chat.Public(fmt.Sprintf(text_DEFUSE_SUCCESS, p.nick))
		g.Stop()
		return

	case defuse_NOTHING:
		g.chat.Public(fmt.Sprintf(text_DEFUSE_NOTHING, p.nick))

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

	g.state = state_PLAYING
	return

}

// Stop ends the game, bomb explodes or turns out to be fake.
func (g *Game) Stop() {

	if g.state != state_PLAYING && g.state != state_DEFUSE {
		return
	}

	// The game ended!
	g.ended = time.Now()
	g.state = state_ENDED

	if !g.bomb.defused {

		g.bomb.location.dead = true

		// Check if the bomb was lying on the ground at detonation time.
		if g.bomb.location == nil {

			g.chat.Public(text_BOMB_EXPLODE)

		} else {

			// Show message and kick players if we can.
			if g.bomb.fake {
				g.chat.Public(text_BOMB_FAKE)
			} else if !tweak_KICK || !g.chat.IsOperator() {
				g.chat.Public(text_BOMB_EXPLODE)
				g.chat.Public(fmt.Sprintf(text_BOMB_EXPLODE_NOOP, g.bomb.location.nick))
			} else {

				if tweak_BAN && g.chat.Ban(g.bomb.location.nick) {

					// Schedule unban!
					go func() {
						timer := time.NewTimer(tweak_BAN_TIME * time.Second)

						// Wait for timer to expire
						<-timer.C

						g.chat.UnBan(g.bomb.location.nick)
					}()
				}

				g.chat.Kick(g.bomb.location.nick, text_BOMB_EXPLODE)
			}

		}

	}

	write(analyse(g))

	g.stop <- true

}

// Leave removes a player after he has left the room.
func (g *Game) Leave(nick string) {

	if g.state == state_INIT || g.state == state_ENDED {
		return
	}

	nick = sanitizeNick(nick)

	// Attempt to fetch the player
	p, playing := g.players[nick]

	// This one isn't playing, couldn't care less.
	if !playing {
		return
	}

	g.chat.Public(fmt.Sprintf(text_PLAYER_LEFT, p.nick))

	delete(g.players, nick)

}

// Rename changes the name of a player.
func (g *Game) Rename(old, nick string) {

	if g.state == state_INIT || g.state == state_ENDED {
		return
	}

	olds := sanitizeNick(old)

	// Attempt to fetch the player
	p, playing := g.players[olds]

	// This one isn't playing, couldn't care less.
	if !playing {
		return
	}

	p.nick = nick
	p.sanitizedNick = sanitizeNick(nick)

	delete(g.players, olds)
	g.players[p.sanitizedNick] = p

	g.chat.Public(fmt.Sprintf(text_PLAYER_RENAME, old, p.nick))

}

// Players shows a list of current players in the channel.
func (g *Game) Players() {

	if g.state != state_PLAYING && g.state != state_DEFUSE {
		return
	}

	players := make([]string, len(g.players))

	i := 0
	for _, player := range g.players {
		players[i] = player.nick
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
		g.Players()

	case cmd_PICK_UP:
		g.Pickup(sender)

	}

}
