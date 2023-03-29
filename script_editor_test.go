package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatScript1(t *testing.T) {
	have := ParseScript("(ally-visible?)")
	expect := "(ally-visible?)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(and 1 2)")
	expect = "(and 1\n     2)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(move (and 1 2))")
	expect = "(move (and 1\n           2))\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(and (and 1 2) 3)")
	expect = "(and (and 1\n          2)\n     3)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(if (or 1 2) (and 2 3) 4)")
	expect = "(if (or 1\n        2)\n  (and 2\n       3)\n  4)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
 	assert.Equal(t, expect, FormatScript(have))
}

func TestCountExpressions(t *testing.T) {
	foo := ParseScript("(if 1 2 3)")
	bar := ParseScript("(if (and 1 2) 3 4)")

	assert.Equal(t, 4, countExpressions(foo))
	assert.Equal(t, 6, countExpressions(bar))
}
