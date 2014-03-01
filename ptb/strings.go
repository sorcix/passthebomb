package ptb

const (

	//
	// START
	//

	// Public; Game start message
	text_START_ATTENTION = "Attentiooooon recruits!"

	// Public; Call to action
	text_START_JOIN = "We have this little b-thing we need taken care of. Type !join to enlist!"

	// Public; Not enough players during join period.
	text_START_FAIL = "Recruits, we need more men! This room is full of pussies!"

	// Public; Game started
	text_START_GO = "GO RECRUITS! Here %s, take the bomb! I have some euhm.. plans to discuss at the bar."

	//
	// END
	//

	// Public; player held the bomb the longest. (%s = winner nickname)
	text_END_WINNER = "Congratulations %s! You've won this round!"

	//
	// HELP (shown during join)
	//

	// Public; Explain how to throw a bomb.
	text_HELP_THROW = "If you have the bomb, use " + cmd_PREFIX + cmd_PASS + " <nick> to throw it to someone else."

	// Public; Explain scoring.
	text_HELP_SCORE = "The longer you hold the bomb, the more points you'll get."

	// Public; Explain how to defuse the bomb.
	text_HELP_DEFUSE = "You can also attempt to defuse the bomb. Type " + cmd_PREFIX + cmd_DEFUSE + " while you have it."

	// Public; The game will start soon. (%s = nickname of first player)
	text_HELP_START = "Prepare yourselves! %s has volunteered to get the bomb first."

	//
	// PLAYERS
	//

	// Public; Player list
	text_PLAYER_LIST = "Platoon: %s"

	// Private; Player joins
	text_PLAYER_JOINED = "You have been enlisted!"

	// Public; Player joins late (%s = nickname)
	text_PLAYER_JOINED_LATE = "Attention platoon! %s joined the game late!"

	// Public; Player wants to leave (%s = nickname)
	text_PLAYER_STOPS = "%s, you pussy! We do not leave the war, soldier."

	// Public; Player changed name. (first %s = old nickname, second %s = new nickname)
	text_PLAYER_RENAME = "Recruits, %s is acting like a complete asshole and is now known as %s!"

	// Public; Player left. (%s = nickname)
	text_PLAYER_LEFT = "DESERTER! %s has gone AWOL!"

	//
	// GAMEPLAY
	//

	// Public; Player throws to himself (%s = nickname)
	text_BOMB_THROWN_SELF = "Recruit %s, stop playing with yourself!"

	// Public; Player throws bomb (first %s = source; second %s = target)
	text_BOMB_THROWN = "%s throws the bomb to %s!"

	// Public; Bomb dropped (%s = wrong target)
	text_BOMB_DROPPED = "No! %s is not in your team, recruit! BOMB DROPPED! (" + cmd_PREFIX + cmd_PICK_UP + ")"

	// Public; Player has picked up a dropped bomb. (%s = nickname)
	text_BOMB_PICKED_UP = "%s has picked up the bomb!"

	// Public; Kick message when the bomb explodes
	text_BOMB_EXPLODE = "beep beep beep beeeeeeeeeep *BOOOOOOOM*"

	// Public; Bomb was fake.
	text_BOMB_FAKE = "beep beep beep beeeeeep... tssss.. ssssh.. [Fake Bomb!]"

	// Public; Bomb explodes but bot can't kick the player. (%s = nick)
	text_BOMB_EXPLODE_NOOP = "The bomb exploded in %s's face!"

	// Public; Bomb sounds, long time.
	text_BOMB_SOUND_LONG = "[BOMB] tsssssss..."

	// Public; Bomb sounds, medium time.
	text_BOMB_SOUND_MEDIUM = "[BOMB] tsssssssSSSSHHH *CRACK*"

	// Public; Bomb sounds, close to detonation..
	text_BOMB_SOUND_SHORT = "[BOMB] BEEP BEEP BEEP BEEP"

	//
	// DEFUSE
	//

	// Public; Player already tried to defuse.. (%s = nickname)
	text_DEFUSE_TRIED = "Sorry %s, you've had your chance! We won't let you mess up twice!"

	// Public; Selected a wire that doesn't exist. (%d = wire number)
	text_DEFUSE_ERROR = "You idiot! There is no wire %d! Don't they learn you how to count these days?"

	// Public; Can't defuse (%s = nickname)
	text_DEFUSE_DISABLED = "Sorry %s, it seems to be impossible to defuse this bomb."

	// Public; Defuse info message
	text_DEFUSE = "Feeling lucky, Cadet? There are %d wires, which one would you like to cut? (" + cmd_PREFIX + cmd_CUT + " <number>)"

	// Public; The wire was already cut by another player.
	text_DEFUSE_DUPLICATE = "You idiot! This wire was already cut.."

	// Public; Timer went down.. Bomb will explode in the next minute.
	text_DEFUSE_LESS_TIME = "Oh no.. You're still alive, but the timer is going crazy! Bomb is unstable!!"

	// Public; Timer went up..
	text_DEFUSE_MORE_TIME = "Phew.. The timer seems to have gone up a bit."

	// Public; Nothing happens (%s = nickname)
	text_DEFUSE_NOTHING = "Nothing happened! Can't you do anything right, %s?"

	// Public; Player defused (%s = nickname)
	text_DEFUSE_SUCCESS = "Amazing! Private %s defused the bomb, you deserve a medal!"

	//
	// COMMANDS
	//

	// Prefix for all game commands
	cmd_PREFIX = "!"

	// Join the game during join phase
	cmd_JOIN = "join"

	// Pass the bomb to someone else
	cmd_PASS = "pass"

	// Start a defuse attempt.
	cmd_DEFUSE = "defuse"

	// Cut a wire during defuse.
	cmd_CUT = "cut"

	// Pick up the bomb when it's on the ground
	cmd_PICK_UP = "pickup"

	// Player list
	cmd_PLAYER_LIST = "players"
)
