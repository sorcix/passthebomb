# Pass The Bomb

Pass The Bomb is a text-based chat game where players pass a bomb around till it explodes. The player holding the bomb when it explodes loses, the one that held it for the longest time wins.

## Usage

1. Install pass the bomb using `go get`:

		go get github.com/sorcix/passthebomb/ptb

2. Implement the `Chat` interface for the chat protocol you're using:

		type Chat interface {
			Public(message string) // Sends a message to all players.
			Private(nick, message string) // Sends a message to given player.
			Kick(nick, reason string) // Kicks a player.
			IsOperator() bool // True if the bot has the power to kick people.
			Ban(nick string) bool // Ban given nickname.
			UnBan(nick string) // Unban given nickname.
		}

4. Create a game instance

		g := ptb.NewGame(chat_interface)

3. Listen for commands in your chatroom and call the game functions:

	* Start
	* Join
	* Throw
	* Pickup
	* Defuse
	* Cut
	* PlayerList
