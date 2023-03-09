package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatScript1(t *testing.T) {
	have := ParseScript("(if 1 2 3)")
	expect := `
(if 1
  2
  3)`
	assert.Equal(t, expect, NewScriptEditor().FormatScript(have, 0))
}

func TestCountExpressions(t *testing.T) {
	foo := ParseScript("(if 1 2 3)")
	bar := ParseScript("(if (and 1 2) 3 4)")

	assert.Equal(t, 4, countExpressions(foo))
	assert.Equal(t, 6, countExpressions(bar))
}
