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
	Actor *Bot
	Target *Cell
}

type Bot struct {
	Team Team
	Id int
	Position *Cell
	Script Script
	Alive bool
}

type GameState struct {
	Arena *Arena
	Bots []Bot
	CurrentBot *Bot
	Tick int
}

type Match struct {
	State *GameState
	Visualizer Visualizer
	Id int
	ScriptA int
	ScriptB int
	ScoreA int
	ScoreB int
}

var currentMatch *Match

func NewMatch(arena *Arena, visualizer Visualizer, id int, scriptId_A int, scriptId_B int) *Match {
	arena.Reset()
	state := &GameState{arena, make([]Bot, BOTS_PER_TEAM * 2), nil, 0}
	match := &Match{state, visualizer, id,  scriptId_A, scriptId_B, 0, 0}

	scriptA := fileManager.LoadScript(state, scriptId_A)
	scriptB := fileManager.LoadScript(state, scriptId_B)

	for teamAindex := 0; teamAindex < BOTS_PER_TEAM; teamAindex++ {
		state.Bots[teamAindex].Team = TeamA
		state.Bots[teamAindex].Id = teamAindex
		state.Bots[teamAindex].Position = arena.TeamASpawns[teamAindex]
		state.Bots[teamAindex].Script = scriptA
		state.Bots[teamAindex].Alive = true

		teamBindex := teamAindex + BOTS_PER_TEAM
		state.Bots[teamBindex].Team = TeamB
		state.Bots[teamBindex].Id = teamBindex
		state.Bots[teamBindex].Position = arena.TeamBSpawns[teamAindex]
		state.Bots[teamBindex].Script = scriptB
		state.Bots[teamBindex].Alive = true
	}

	return match
}

// Returns true if the game is over and false if it's still going.
func (m *Match) RunTick() bool {
	for i := range m.State.Bots { // FIXME: Alternate between teams
		m.RunOneBot(&m.State.Bots[i])
	}
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
	m.State.Tick++
	return false
}

func (m *Match) RunOneBot(bot *Bot) {
	m.State.CurrentBot = bot
	action := bot.Script.Run().Action
	switch action.Type {
	case ActionWait:
		bot.Position.Waits++
	case ActionMove:
		m.BotMove(bot, action.Target)
	}
	//   update cell statistics?
	//   update score?
}

// If the space is passable but another bot is in the space, it's the same as hitting a wall.
func (m *Match) BotMove(bot *Bot, destination *Cell) {
	for _, otherBot := range m.State.Bots {
		if bot.Id != otherBot.Id && bot.Position == destination {
			destination = bot.Position
		}
	}

	if bot.Position != destination {
		destination.Moves++
	}

	bot.Position = destination
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
