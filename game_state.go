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
		if bot.Position == cell && bot.Alive {
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

func (gs *GameState) FirstNonEmptyCellOnLine(src *Cell, dest *Cell) *Cell {
	var found *Cell = nil

	BresenhamLine(src.X, src.Y, dest.X, dest.Y, func(x, y int) bool {
		c := &gs.Arena.Cells[x * gs.Arena.Height + y]
		if !gs.CellIsEmpty(c) && c != src {
			found = c
			return false
		}
		return true
	})
	return found
}

func (gs *GameState) NearestVisibleEnemyOrGoal() *Cell {
	closestDistance := gs.Arena.Width * 100
	var closestTarget *Cell = nil

	for i := range gs.Bots {
		bot := &gs.Bots[i]
		if gs.CurrentBot.Team != bot.Team && bot.Alive && gs.Arena.CanSee(gs.CurrentBot.Position, bot.Position) {
			distance := gs.Arena.Distance(gs.CurrentBot.Position, bot.Position)
			if distance < closestDistance {
				closestDistance = distance
				closestTarget = bot.Position
			}
		}
	}

	for i := range gs.Goals {
		goal := &gs.Goals[i]
		if gs.CurrentBot.Team != goal.Team && goal.Alive && gs.Arena.CanSee(gs.CurrentBot.Position, goal.Position) {
			distance := gs.Arena.Distance(gs.CurrentBot.Position, goal.Position)
			if distance < closestDistance {
				closestDistance = distance
				closestTarget = goal.Position
			}
		}
	}

	return closestTarget
}

func (gs *GameState) CountVisibleEnemies() int {
	count := 0

	for i := range gs.Bots {
		bot := &gs.Bots[i]
		if gs.CurrentBot.Team != bot.Team && bot.Alive && gs.Arena.CanSee(gs.CurrentBot.Position, bot.Position) {
			count++
		}
	}
	return count
}

func (gs *GameState) GoalVisible(team Team) bool {
	for _, goal := range gs.Goals {
		if goal.Team == team && gs.Arena.CanSee(gs.CurrentBot.Position, goal.Position) {
			return true
		}
	}
	return false
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
		logger.Printf("A team died: A %d, B %d", alive[TeamA], alive[TeamB])
		return true
	}
	if !gs.Goals[TeamA].Alive || !gs.Goals[TeamB].Alive { // A goal has been destroyed
		logger.Printf("A goal died: A %v, B %v", gs.Goals[TeamA].Alive, gs.Goals[TeamB].Alive)
		return true
	}
	if gs.Tick >= MAX_TICKS_PER_GAME { // The game has run over the max allowed time
		logger.Printf("Game ran out of time.")
		return true
	}

	return false
}
