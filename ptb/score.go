package ptb

import (
	"time"
)

// ScoreCalc represents a function able to calculate the score for a player.
type ScoreCalc func(g *Game, p *Player) *ScoreCard

// ScoreCard represents a single player on the scoreboard.
type ScoreCard struct {
	Player *Player
	Score  uint64
}

// ScoreBoard represents the list of players and their scores.
type ScoreBoard []*ScoreCard

func (sb ScoreBoard) Len() int           { return len(sb) }
func (sb ScoreBoard) Less(i, j int) bool { return sb[i].Score < sb[j].Score }
func (sb ScoreBoard) Swap(i, j int)      { sb[i], sb[j] = sb[j], sb[i] }

// DurationScore scores players based on the time they've held the bomb.
// Score is given in seconds.
func DurationScore(g *Game, p *Player) (c *ScoreCard) {

	c = new(ScoreCard)
	c.Player = p

	if p.Dead {
		c.Score = 0
		return
	}

	c.Score = uint64(p.Duration / time.Second)
	return
}

// MeanDurationScore scores players based on the mean time they've held the bomb per turn.
// Score is given in seconds.
func MeanDurationScore(g *Game, p *Player) (c *ScoreCard) {

	c = new(ScoreCard)
	c.Player = p

	if p.Dead {
		c.Score = 0
		return
	}

	c.Score = uint64(p.MeanDuration / time.Second)
	return
}

// DefuseScore scores players based on the time they've held the bomb, with a bonus for defuse.
// Score is given in seconds.
func DefuseScore(g *Game, p *Player) (c *ScoreCard) {

	c = new(ScoreCard)
	c.Player = p

	if p.Dead {
		c.Score = 0
		return
	}

	c.Score = uint64(p.Duration / time.Second)

	// The player that defused the bomb gets 5 minutes bonus!
	if p.Defused {
		c.Score = c.Score + 60*5
	}

	return
}

// ComplexScore tries to score players using as much statistics as possible.
// Score is given in seconds.
func ComplexScore(g *Game, p *Player) (c *ScoreCard) {

	c = new(ScoreCard)
	c.Player = p

	if p.Dead {
		c.Score = 0
		return
	}

	c.Score = uint64(p.Duration / time.Second)

	// Trying to defuse is worth a minute bonus.
	if p.DefuseAttempt {
		c.Score = c.Score + 60
	}

	// The player that defused the bomb gets 5 minutes bonus!
	if p.Defused {
		c.Score = c.Score + 60*5
	}

	// Bonus time if the player held the bomb for a longer time.
	if p.MeanDuration > time.Minute {
		c.Score = c.Score + 60*5
	}

	return

}
