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
	expect = "(and 1 2)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(and 1 (+ 2 2))")
	expect = "(and 1\n     (+ 2 2))\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(move (and 1 2))")
	expect = "(move (and 1 2))\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(move (and 1 (+ 2 2)))")
	expect = "(move (and 1\n           (+ 2 2)))\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(and (and 1 2) 3)")
	expect = "(and (and 1 2)\n     3)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(and (and 1 (+ 2 2)) 3)")
	expect = "(and (and 1\n          (+ 2 2))\n     3)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
	assert.Equal(t, expect, FormatScript(have))

	have = ParseScript("(if (or 1 2) (and 2 3) 4)")
	expect = "(if (or 1 2)\n  (and 2 3)\n  4)\n"
	// fmt.Printf("have:\n%s\nexpect:\n%s\n", FormatScript(have), expect)
 	assert.Equal(t, expect, FormatScript(have))
}

func TestSimplifyScript(t *testing.T) {
	tests := map[string]string{
		"(if 1 (+ 2 2) (- 3 3))": "4\n",
		"(and (- 2 2) (+ 3 3))": "0\n",
		"(and (+ 2 2) (+ 3 3))": "6\n",
		"(* 2 0)": "0\n",
		"(* 2 6)": "12\n",
		"(/ 2 0)": "0\n",
		"(/ 16 2)": "8\n",
		"(mod 2 0)": "0\n",
		"(mod 16 5)": "1\n",
		"(or 1 (my-x-pos))": "1\n",
		"(or 0 (/ 2 0))": "0\n",
		"(< (+ 2 2) (+ 3 3))": "1\n",
		"(< (+ 3 3) (+ 2 2))": "0\n",
		"(> (+ 2 2) (+ 3 3))": "0\n",
		"(> (+ 3 3) (+ 2 2))": "1\n",
		"(= (+ 2 2) (+ 2 2))": "1\n",
		"(= (+ 3 3) (+ 2 2))": "0\n",
		"(not (+ 2 2))": "0\n",
		"(not (- 2 2))": "1\n",
		"(shoot (+ 2 2))": "(shoot 4)\n",


		// Expressions that are not constant should not be simplified.
		"(if (my-x-pos) (+ 2 2) (- 3 3))": "(if (my-x-pos)\n  4\n  0)\n",
		"(and (+ 2 (my-x-pos)) (+ 3 3))": "(and (+ 2 (my-x-pos))\n     6)\n",
		"(or (my-x-pos) (my-x-pos))": "(or (my-x-pos) (my-x-pos))\n",
	}

	for before, after := range tests {
		code := ParseScript(before)
		SimplifyTree(code)
		assert.Equal(t, after, FormatScript(code))
	}
}
