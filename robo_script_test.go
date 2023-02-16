package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScript1(t *testing.T) {
	script := "(+ 1 32)"
	node, remaining, err := readToken(script)

	assert.Empty(t, remaining)
	assert.NoError(t, err)

	assert.Equal(t, Expr, node.Type)
	assert.Equal(t, 3, len(node.Children))

	assert.Empty(t, node.Children[0].Children)
	assert.Equal(t, Symbol, node.Children[0].Type)
	assert.Equal(t, "+", node.Children[0].Sym)

	assert.Empty(t, node.Children[1].Children)
	assert.Equal(t, Int, node.Children[1].Type)
	assert.Equal(t, 1, node.Children[1].N)

	assert.Empty(t, node.Children[2].Children)
	assert.Equal(t, Int, node.Children[2].Type)
	assert.Equal(t, 32, node.Children[2].N)
}

func TestScript2(t *testing.T) {
	script := "(+ 1 (foo 22))"
	node, remaining, err := readToken(script)

	assert.Empty(t, remaining)
	assert.NoError(t, err)

	assert.Equal(t, Expr, node.Type)
	assert.Equal(t, 3, len(node.Children))

	assert.Empty(t, node.Children[0].Children)
	assert.Equal(t, Symbol, node.Children[0].Type)
	assert.Equal(t, "+", node.Children[0].Sym)

	assert.Empty(t, node.Children[1].Children)
	assert.Equal(t, Int, node.Children[1].Type)
	assert.Equal(t, 1, node.Children[1].N)

	assert.Equal(t, 2, len(node.Children[2].Children))
	assert.Equal(t, Expr, node.Children[2].Type)

	assert.Empty(t, node.Children[2].Children[0].Children)
	assert.Equal(t, Symbol, node.Children[2].Children[0].Type)
	assert.Equal(t, "foo", node.Children[2].Children[0].Sym)

	assert.Empty(t, node.Children[2].Children[1].Children)
	assert.Equal(t, Int, node.Children[2].Children[1].Type)
	assert.Equal(t, 22, node.Children[2].Children[1].N)
}

func TestScriptUnterminatedError(t *testing.T) {
	script := "(+ 1 2"
	node, _, err := readToken(script)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Unterminated expression")
}

func TestScriptIntInFunctionPosition(t *testing.T) {
	script := "(1 + 2)"
	node, _, err := readToken(script)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Non-symbol in function position")
}

func TestScriptExprInFunctionPosition(t *testing.T) {
	script := "((+ 1 2) 3)"
	node, _, err := readToken(script)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Non-symbol in function position")
}

func TestScriptEmptyList(t *testing.T) {
	script := "(+ 1 ())"
	node, _, err := readToken(script)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Found an empty list")
}

func TestScriptSymbolInArgumentPosition(t *testing.T) {
	script := "(+ foo 2)"
	node, _, err := readToken(script)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Symbol 'foo' passed as function argument")
}

func TestAddNumbers(t *testing.T) {
	script := "(+ 13 2)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 15, result.Int)
}

func TestIf(t *testing.T) {
	script := "(if 4 1 2)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)

	script = "(if 0 1 2)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)
}

func TestLessThan(t *testing.T) {
	script := "(< 1 2)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)

	script = "(< 2 1)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)
}

func TestGreaterThan(t *testing.T) {
	script := "(> 1 2)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	script = "(> 2 1)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)
}

func TestEqual(t *testing.T) {
	script := "(= 1 2)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	script = "(= 2 2)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)
}

func TestAnd(t *testing.T) {
	script := "(and 1 2)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)

	script = "(and 0 2)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	script = "(and 2 0)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)
}

func TestOr(t *testing.T) {
	script := "(or 0 0)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	script = "(or 0 2)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)

	script = "(or 2 0)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)
}

func TestNot(t *testing.T) {
	script := "(not 0)"
	node, _, err := readToken(script)
	assert.NoError(t, err)

	result := node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)

	script = "(not 33)"
	node, _, err = readToken(script)
	assert.NoError(t, err)

	result = node.Eval()
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)
}

