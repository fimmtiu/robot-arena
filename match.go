package main

import (
	"math/rand"
)

type Direction int
const (
	North Direction = iota
	South
	East
	West
	NumberOfDirections
)

type ActionType int
const (
	ActionMove ActionType = iota
	ActionWait
	ActionShoot
)

type Action struct {
	Type ActionType
	Target *Cell
}

type Bot struct {
	Team Team
	Id int
	Position *Cell
	Script Script
	Alive bool
}

type Goal struct {
	Team Team
	Position *Cell
	Alive bool
}

type Match struct {
	Rand *rand.Rand
	State *GameState
	Generation *Generation
	Id int
	ScriptA int
	ScriptB int
	Scores [2]int
}

var currentMatch *Match
var turnSequence = []int{0, 5, 1, 6, 2, 7, 3, 8, 4, 9}  // Alternates bots from different teams

func NewMatch(generation *Generation, id int, scriptId_A int, scriptId_B int) *Match {
	generation.Arena.Reset()
	rng := rand.New(rand.NewSource(int64(id)))
	state := NewGameState(generation.Arena)
	match := &Match{rng, state, generation, id,  scriptId_A, scriptId_B, [2]int{0, 0}}

	scripts := [2]Script{generation.FileManager.LoadScript(state, scriptId_A), generation.FileManager.LoadScript(state, scriptId_B)}
	for i, bot := range state.Bots {
		state.Bots[i].Script = scripts[bot.Team]
	}

	generation.Visualizer.Init(state)
	return match
}

func (m *Match) Run() {
	for {
		if m.RunTick() {   // Returns true when the match is over
			break
		}
	}
}

// Returns true if the game is over and false if it's still going.
func (m *Match) RunTick() bool {
	for _, i := range turnSequence {
		if m.State.Bots[i].Alive {
			m.RunOneBot(&m.State.Bots[i])
		} else {
			m.Generation.Visualizer.NoChange()
		}
	}

	m.Generation.Visualizer.TickComplete()
	m.State.Tick++
	if m.State.Tick >= MAX_TICKS_PER_GAME {  // Penalize both teams if the game runs too long.
		m.Scores[TeamA] -= 5
		m.Scores[TeamB] -= 5
	}

	if m.State.IsGameOver() {
		logger.Printf("Final score: Team A %d, Team B %d.", m.Scores[TeamA], m.Scores[TeamB])
		m.Generation.Visualizer.Finish()
		return true
	}
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
	case ActionShoot:
		m.BotShoot(bot, action)
	}

	m.Generation.Visualizer.Update(action)
}

// If the space is passable but another bot is in the space, it's the same as hitting a wall.
func (m *Match) BotMove(bot *Bot, destination *Cell) {
	if m.State.CellIsEmpty(destination) {
		bot.Position = destination
		destination.Moves++
	}
}

// We shoot at the first non-empty cell along the line between you and the target.
func (m *Match) BotShoot(bot *Bot, action Action) {
	// We modify the action so that the visualizer will later know which target was actually hit.
	action.Target = m.State.FirstNonEmptyCellOnLine(bot.Position, action.Target)


	// Accuracy falls off pretty severely with distance. Will need to adjust this eventually.
	hitChance := 1.0 - (float32(m.State.Arena.Distance(bot.Position, action.Target)) * 0.03)
	if m.Rand.Float32() <= hitChance {
		targetBot := m.State.BotAtCell(action.Target)
		targetGoal := m.State.GoalAtCell(action.Target)

		if targetBot != nil {
			targetBot.Alive = false
			targetBot.Position.Kills++
			if targetBot.Team == bot.Team {
				logger.Printf("Friendly fire on team %d! Bot %d killed bot %d. (%d, %d)", bot.Team, bot.Id, targetBot.Id, targetBot.Position.X, targetBot.Position.Y)
				m.Scores[bot.Team] -= 2  // penalty for friendly fire
			} else {
				logger.Printf("Bot %d from team %d killed enemy bot %d", bot.Id, bot.Team, targetBot.Id)
				m.Scores[bot.Team] += 1
			}
		} else if targetGoal != nil {
			targetGoal.Alive = false
			if targetGoal.Team == bot.Team {
				logger.Printf("Own goal for team %d!", bot.Team)
				m.Scores[bot.Team] -= 20  // massive penalty for an own-goal
			} else {
				logger.Printf("Team %d destroyed the other team's goal", bot.Team)
				m.Scores[bot.Team] += 10
			}
		}
		// Otherwise you probably shot a wall, so we do nothing.
	}

	bot.Position.Shots++
}
