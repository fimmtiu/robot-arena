package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScript1(t *testing.T) {
	code := "(+ 1 32)"
	node, remaining, err := readToken(code)

	assert.Empty(t, remaining)
	assert.NoError(t, err)

	assert.Equal(t, Expr, node.Type)
	assert.Equal(t, 3, len(node.Children))

	assert.Empty(t, node.Children[0].Children)
	assert.Equal(t, FuncName, node.Children[0].Type)
	assert.Equal(t, "+", node.Children[0].Func.Name)

	assert.Empty(t, node.Children[1].Children)
	assert.Equal(t, Int, node.Children[1].Type)
	assert.Equal(t, 1, node.Children[1].N)

	assert.Empty(t, node.Children[2].Children)
	assert.Equal(t, Int, node.Children[2].Type)
	assert.Equal(t, 32, node.Children[2].N)
}

func TestScript2(t *testing.T) {
	code := "(+ 1 (not 22))"
	node, remaining, err := readToken(code)

	assert.Empty(t, remaining)
	assert.NoError(t, err)

	assert.Equal(t, Expr, node.Type)
	assert.Equal(t, 3, len(node.Children))

	assert.Empty(t, node.Children[0].Children)
	assert.Equal(t, FuncName, node.Children[0].Type)
	assert.Equal(t, "+", node.Children[0].Func.Name)

	assert.Empty(t, node.Children[1].Children)
	assert.Equal(t, Int, node.Children[1].Type)
	assert.Equal(t, 1, node.Children[1].N)

	assert.Equal(t, 2, len(node.Children[2].Children))
	assert.Equal(t, Expr, node.Children[2].Type)

	assert.Empty(t, node.Children[2].Children[0].Children)
	assert.Equal(t, FuncName, node.Children[2].Children[0].Type)
	assert.Equal(t, "not", node.Children[2].Children[0].Func.Name)

	assert.Empty(t, node.Children[2].Children[1].Children)
	assert.Equal(t, Int, node.Children[2].Children[1].Type)
	assert.Equal(t, 22, node.Children[2].Children[1].N)
}

func TestScriptUnterminatedError(t *testing.T) {
	code := "(+ 1 2"
	node, _, err := readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Unterminated expression")
}

func TestScriptIntInFunctionPosition(t *testing.T) {
	code := "(1 + 2)"
	node, _, err := readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Non-symbol in function position")
}

func TestScriptExprInFunctionPosition(t *testing.T) {
	code := "((+ 1 2) 3)"
	node, _, err := readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Non-symbol in function position")
}

func TestScriptEmptyList(t *testing.T) {
	code := "(+ 1 ())"
	node, _, err := readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Found an empty list")
}

func TestScriptSymbolInArgumentPosition(t *testing.T) {
	code := "(+ not 2)"
	node, _, err := readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Symbol 'not' passed as function argument")
}

func TestUndefinedFunction(t *testing.T) {
	code := "(monkey 1 2)"
	node, _, err := readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "No such function: 'monkey'")

}

func TestArityErrors(t *testing.T) {
	code := "(+ 1 2 3)"
	node, _, err := readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Wrong number of arguments to '+': got 3, expected 2")

	code = "(+ 1)"
	node, _, err = readToken(code)

	assert.Nil(t, node)
	assert.ErrorContains(t, err, "Wrong number of arguments to '+': got 1, expected 2")
}

func TestScriptNodeSize(t *testing.T) {
	foo := ParseScript("(if 1 2 3)")
	bar := ParseScript("(if (and 1 2) 3 4)")

	assert.Equal(t, 4, foo.Size())
	assert.Equal(t, 6, bar.Size())
}

func TestAddNumbers(t *testing.T) {
	code := "(+ 13 2)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 15, result.Int)
}

func TestIf(t *testing.T) {
	code := "(if 4 1 2)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)

	code = "(if 0 1 2)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)
}

func TestLessThan(t *testing.T) {
	code := "(< 1 2)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)

	code = "(< 2 1)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)
}

func TestGreaterThan(t *testing.T) {
	code := "(> 1 2)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	code = "(> 2 1)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)
}

func TestEqual(t *testing.T) {
	code := "(= 1 2)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	code = "(= 2 2)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)
}

func TestAnd(t *testing.T) {
	code := "(and 1 2)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)

	code = "(and 0 2)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	code = "(and 2 0)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)
}

func TestOr(t *testing.T) {
	code := "(or 0 0)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)

	code = "(or 0 2)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)

	code = "(or 2 0)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 2, result.Int)
}

func TestNot(t *testing.T) {
	code := "(not 0)"
	node, _, err := readToken(code)
	assert.NoError(t, err)

	script := Script{nil, nil}
	result := script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 1, result.Int)

	code = "(not 33)"
	node, _, err = readToken(code)
	assert.NoError(t, err)

	result = script.Eval(node)
	assert.Equal(t, ResultInt, result.Type)
	assert.Equal(t, 0, result.Int)
}

