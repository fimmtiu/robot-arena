package main

type GameState struct {
	Arena *Arena
	Bots []Bot
	Goals [2]Goal
	CurrentBot *Bot
	Tick int
}

func NewGameState(arena *Arena) *GameState {
	state := &GameState{arena, make([]Bot, BOTS_PER_TEAM * 2), [2]Goal{}, nil, 0}

	// Team A occupies slots 0-4. Team B occupies slots 5-9.
	for teamAindex := 0; teamAindex < BOTS_PER_TEAM; teamAindex++ {
		state.Bots[teamAindex].Team = TeamA
		state.Bots[teamAindex].Id = teamAindex
		state.Bots[teamAindex].Position = arena.Spawns[TeamA][teamAindex]
		state.Bots[teamAindex].Alive = true

		teamBindex := teamAindex + BOTS_PER_TEAM
		state.Bots[teamBindex].Team = TeamB
		state.Bots[teamBindex].Id = teamBindex
		state.Bots[teamBindex].Position = arena.Spawns[TeamB][teamAindex]
		state.Bots[teamBindex].Alive = true
	}

	state.Goals[TeamA] = Goal{Team: TeamA, Position: arena.Goals[TeamA], Alive: true}
	state.Goals[TeamB] = Goal{Team: TeamB, Position: arena.Goals[TeamB], Alive: true}

	return state
}

func (gs *GameState) BotAtCell(cell *Cell) *Bot {
	for i, bot := range gs.Bots {
		if bot.Position == cell {
			return &gs.Bots[i]
		}
	}
	return nil
}

func (gs *GameState) GoalAtCell(cell *Cell) *Goal {
	for i, goal := range gs.Goals {
		if goal.Position == cell {
			return &gs.Goals[i]
		}
	}
	return nil
}

func (gs *GameState) CellIsEmpty(cell *Cell) bool {
	return cell.BotsCanPass() && gs.BotAtCell(cell) == nil
}

func (gs *GameState) IsGameOver() bool {
	alive := [2]int{0, 0}
	for _, bot := range gs.Bots {
		if bot.Alive {
			alive[bot.Team]++
		}
	}

	if alive[TeamA] == 0 || alive[TeamB] == 0 { // One team is wiped out
		return true
	}
	if !gs.Goals[TeamA].Alive || !gs.Goals[TeamB].Alive { // A goal has been destroyed
		return true
	}
	if gs.Tick >= MAX_TICKS_PER_GAME { // The game has run over the max allowed time
		return true
	}

	return false
}
