package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateMatchups(t *testing.T) {
	g := &Generation{1, nil, nil, nil, [][2]int{}}
	scriptIds := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	matchesPerScript := 5
	matchCounts := make(map[int]int, len(scriptIds))
	g.calculateMatchups(scriptIds, matchesPerScript)

	logger.Printf("matchups: %v", g.matchups)

	assert.Equal(t, matchesPerScript * len(scriptIds) / 2, len(g.matchups))
	for _, matchup := range g.matchups {
		matchCounts[matchup[0]]++
		matchCounts[matchup[1]]++
		assert.NotEqual(t, matchup[0], matchup[1])
	}

	for id, count := range matchCounts {
		logger.Printf("id: %v, count: %v", id, count)
		assert.True(t, count == matchesPerScript || count == matchesPerScript + 1)
	}
}
