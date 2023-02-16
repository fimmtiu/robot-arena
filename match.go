package main

type Direction int
const (
	North Direction = iota
	South
	East
	West
)

type ActionType int
const (
	ActionMove ActionType = iota
	ActionWait
	ActionShoot
)

type Action struct {
	Type ActionType
	Actor Bot
	Target *Cell
}

type Bot struct {
	Team Team
	Id int
	Position *Cell
	Script *ScriptNode
	Alive bool
}

type Match struct {
	Arena *Arena
	Visualizer Visualizer
	Id int
	Bots []Bot
	Tick int
	ScriptA int
	ScriptB int
	ScoreA int
	ScoreB int
}

var currentMatch *Match

func NewMatch(arena *Arena, visualizer Visualizer, id int, scriptId_A int, scriptId_B int) *Match {
	arena.Reset()
	match := &Match{
		arena, visualizer, id, make([]Bot, BOTS_PER_TEAM * 2), 0, scriptId_A, scriptId_B, 0, 0,
	}

	scriptA := fileManager.LoadScript(scriptId_A)
	scriptB := fileManager.LoadScript(scriptId_B)

	for teamAindex := 0; teamAindex < BOTS_PER_TEAM; teamAindex++ {
		match.Bots[teamAindex].Team = TeamA
		match.Bots[teamAindex].Id = teamAindex
		match.Bots[teamAindex].Position = arena.TeamASpawns[teamAindex]
		match.Bots[teamAindex].Script = scriptA
		match.Bots[teamAindex].Alive = true

		teamBindex := teamAindex + BOTS_PER_TEAM
		match.Bots[teamBindex].Team = TeamB
		match.Bots[teamBindex].Id = teamBindex
		match.Bots[teamBindex].Position = arena.TeamBSpawns[teamAindex]
		match.Bots[teamBindex].Script = scriptB
		match.Bots[teamBindex].Alive = true
	}

	return match
}

// Returns true if the game is over and false if it's still going.
func (m *Match) RunTick() bool {
	// for each bot
	//   run its script
	//   get associated action
	//   do it
	//   update cell statistics
	//   update score
	//   trigger the visualizer
	// if the game is over
	//   return true

	if m.Visualizer != nil {
		m.Visualizer.Update(Action{Type: ActionWait})
	}
	m.Tick++
	return false
}

func (m *Match) BotMove(bot Bot, relativeDirection Direction) {
	actualDirection := relativeToActualDirection(relativeDirection, bot.Team)
	destinationCell := m.Arena.DestinationCellAfterMove(bot.Position, actualDirection)

	for _, otherBot := range m.Bots {
		if bot.Id != otherBot.Id && bot.Position == destinationCell {
			destinationCell = bot.Position
		}
	}

	if bot.Position != destinationCell {
		destinationCell.Moves++
	}

	bot.Position = destinationCell
}

// Each team considers "north" to be the direction of the enemy's goal, and "south" to be the direction of its own side.
// This allows a script to run identically regardless of whether it's controlling Team A or Team B.
func relativeToActualDirection(relative Direction, team Team) Direction {
	if team == TeamA {
		switch relative {
		case North: return East
		case South: return West
		case East:  return South
		case West:  return North
		}
	} else {
		switch relative {
		case North: return West
		case South: return East
		case East:  return North
		case West:  return South
		}
	}
	logger.Fatalf("Weird direction %d for team %d!", relative, team)
	return North
}
