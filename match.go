package main

import "math/rand"

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

type Goal struct {
	Team Team
	Position *Cell
	Alive bool
}

type Match struct {
	Rand *rand.Rand
	State *GameState
	Visualizer Visualizer
	Id int
	ScriptA int
	ScriptB int
	Scores [2]int
}

var currentMatch *Match
var turnSequence = []int{0, 5, 1, 6, 2, 7, 3, 8, 4, 9}  // Alternates bots from different teams

func NewMatch(arena *Arena, visualizer Visualizer, id int, scriptId_A int, scriptId_B int) *Match {
	arena.Reset()
	rng := rand.New(rand.NewSource(int64(id)))
	state := NewGameState(arena)
	match := &Match{rng, state, visualizer, id,  scriptId_A, scriptId_B, [2]int{0, 0}}

	scripts := [2]Script{fileManager.LoadScript(state, scriptId_A), fileManager.LoadScript(state, scriptId_B)}
	for i, bot := range state.Bots {
		state.Bots[i].Script = scripts[bot.Team]
	}

	return match
}

// Returns true if the game is over and false if it's still going.
func (m *Match) RunTick() bool {
	for _, i := range turnSequence {
		if m.State.Bots[i].Alive {
			m.RunOneBot(&m.State.Bots[i])
		}
	}

	m.State.Tick++
	if m.State.Tick >= MAX_TICKS_PER_GAME {  // Penalize both teams if the game runs too long.
		m.Scores[TeamA] -= 5
		m.Scores[TeamB] -= 5
	}
	return m.State.IsGameOver()
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
		m.BotShoot(bot, action.Target)
	}

	if m.Visualizer != nil {
		m.Visualizer.Update(action)
	}
}

// If the space is passable but another bot is in the space, it's the same as hitting a wall.
func (m *Match) BotMove(bot *Bot, destination *Cell) {
	if m.State.CellIsEmpty(destination) {
		bot.Position = destination
		destination.Moves++
	}
}

func (m *Match) BotShoot(bot *Bot, target *Cell) {
	if m.Rand.Float32() <= 0.7 {  // FIXME: For now, let's just give them a 70% chance of hitting.
		targetBot := m.State.BotAtCell(target)
		if targetBot != nil {
			targetBot.Alive = false
			targetBot.Position.Kills++
			if targetBot.Team == bot.Team {
				m.Scores[bot.Team] -= 2  // penalty for friendly fire
			} else {
				m.Scores[bot.Team] += 1
			}

		} else {
			targetGoal := m.State.GoalAtCell(target)
			if targetGoal == nil {
				logger.Fatalf("Fired at an empty cell? %v", target)
			}
			targetGoal.Alive = false
			if targetBot.Team == bot.Team {
				m.Scores[bot.Team] -= 20  // massive penalty for an own-goal
			} else {
				m.Scores[bot.Team] += 10
			}
		}
	}

	bot.Position.Shots++
}
