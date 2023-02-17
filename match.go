package main

import "math/rand"

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

type Goal struct {
	Team Team
	Position *Cell
	Alive bool
}

type GameState struct {
	Arena *Arena
	Bots []Bot
	Goals [2]Goal
	CurrentBot *Bot
	Tick int
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
	state := &GameState{arena, make([]Bot, BOTS_PER_TEAM * 2), [2]Goal{}, nil, 0}
	match := &Match{rng, state, visualizer, id,  scriptId_A, scriptId_B, [2]int{0, 0}}

	scriptA := fileManager.LoadScript(state, scriptId_A)
	scriptB := fileManager.LoadScript(state, scriptId_B)

	// Team A occupies slots 0-4. Team B occupies slots 5-9.
	for teamAindex := 0; teamAindex < BOTS_PER_TEAM; teamAindex++ {
		state.Bots[teamAindex].Team = TeamA
		state.Bots[teamAindex].Id = teamAindex
		state.Bots[teamAindex].Position = arena.Spawns[TeamA][teamAindex]
		state.Bots[teamAindex].Script = scriptA
		state.Bots[teamAindex].Alive = true

		teamBindex := teamAindex + BOTS_PER_TEAM
		state.Bots[teamBindex].Team = TeamB
		state.Bots[teamBindex].Id = teamBindex
		state.Bots[teamBindex].Position = arena.Spawns[TeamB][teamAindex]
		state.Bots[teamBindex].Script = scriptB
		state.Bots[teamBindex].Alive = true
	}

	state.Goals[TeamA] = Goal{Team: TeamA, Position: arena.Goals[TeamA], Alive: true}
	state.Goals[TeamB] = Goal{Team: TeamB, Position: arena.Goals[TeamB], Alive: true}

	return match
}

// Returns true if the game is over and false if it's still going.
func (m *Match) RunTick() bool {
	for _, i := range turnSequence {
		if m.State.Bots[i].Alive {
			m.RunOneBot(&m.State.Bots[i])
		}
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

	m.State.Tick++
	return m.IsGameOver()
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
	//   update cell statistics?
	//   update score?

	if m.Visualizer != nil {
		m.Visualizer.Update(action)
	}
}

// If the space is passable but another bot is in the space, it's the same as hitting a wall.
func (m *Match) BotMove(bot *Bot, destination *Cell) {
	for _, otherBot := range m.State.Bots {
		if bot.Id != otherBot.Id && bot.Position == destination {
			destination = bot.Position
		}
	}

	// We only increment the space's move counter if the bot successfully moved, not if it hit something.
	if bot.Position != destination {
		destination.Moves++
	}

	bot.Position = destination
}

func (m *Match) BotShoot(bot *Bot, target *Cell) {
	if m.Rand.Float32() <= 0.7 {  // FIXME: For now, let's just give them a 70% chance of hitting.
		targetBot := m.BotAtCell(target)
		targetBot.Alive = false
		targetBot.Position.Kills++
		if targetBot.Team == bot.Team {
			m.Scores[bot.Team] -= 2  // penalty for friendly fire
		} else {
			m.Scores[bot.Team] += 1
		}
	}

	bot.Position.Shots++
}

func (m *Match) BotAtCell(cell *Cell) *Bot {
	for i, bot := range m.State.Bots {
		if bot.Position == cell {
			return &m.State.Bots[i]
		}
	}
	return nil
}

func (m *Match) IsGameOver() bool {
	alive := [2]int{0, 0}
	for _, bot := range m.State.Bots {
		if bot.Alive {
			alive[bot.Team]++
		}
	}

	if alive[TeamA] == 0 || alive[TeamB] == 0 { // One team is wiped out
		return true
	}
	if !m.State.Goals[TeamA].Alive || !m.State.Goals[TeamB].Alive { // A goal has been destroyed
		return true
	}
	if m.State.Tick >= MAX_TICKS_PER_GAME { // The game has run over the max allowed time
		return true
	}

	return false
}
