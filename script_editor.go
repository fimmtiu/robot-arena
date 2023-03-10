package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

const MIN_EXPRS_PER_SCRIPT = 20  // FIXME: Later, let's use the average size of scripts in the generation instead of a constant
const MUTATIONS_PER_SCRIPT = 2 // should this be random?
const MAX_LINE_LEN = 40
const INTEGER_PERCENT = 0.3  // 30 percent of all randomly generated nodes will be integers.

var oneLineFormatStrings = []string{
	"%s(%s)",
	"%s(%s %s)",
	"%s(%s %s %s)",
	"%s(%s %s %s %s)",
}
var multiLineFormatStrings = []string{
	"%s(%s)",
	"%s(%s %s)",
	"%s(%s %s\n%s %s)",
	"%s(%s %s\n%s %s\n%s %s)",
}

type ScriptEditor struct {
	// FIXME: Will this ever need any state, or should I change these methods to plain functions?
}

func NewScriptEditor() *ScriptEditor {
	return &ScriptEditor{}
}

func (editor *ScriptEditor) RandomScript() string {
	script := editor.randomNode()
	for countExpressions(script) < MIN_EXPRS_PER_SCRIPT {
		script = editor.wrapNode(script)
	}
	return editor.FormatScript(script)
}

func (editor *ScriptEditor) randomNode() *ScriptNode {
	if rand.Float32() < INTEGER_PERCENT {
		return &ScriptNode{Type: Int, N: randomInt()}
	} else {
		randFunction := AllFunctions[rand.Intn(len(AllFunctions))]
		node := &ScriptNode{Type: Expr, Children: []*ScriptNode{{Type: FuncName, Func: randFunction}}}
		for i := 0; i < randFunction.Arity; i++ {
			node.Children = append(node.Children, editor.randomNode())
		}
		return node
	}
}

// A curve that gives us numbers between 0 and 50, with more small numbers (0-5) than large ones.
// https://www.desmos.com/calculator/onchb78rot
func randomInt() int {
	return int(math.Floor(0.00005 * math.Pow(rand.Float64() * 100, 3)))
}

// Counts the number of expressions in a ScriptNode tree.
func countExpressions(node *ScriptNode) int {
	if node.Type == Expr {
		i := 0
		for _, child := range node.Children {
			i += countExpressions(child)
		}
		return i
	} else {
		return 1
	}
}

// Wraps a node in some other multi-argument expression.
func (editor *ScriptEditor) wrapNode(node *ScriptNode) *ScriptNode {
	for {
		fn := AllFunctions[rand.Intn(len(AllFunctions))]
		if fn.Arity > 0 {
			insertAt := rand.Intn(fn.Arity)
			expr := &ScriptNode{Type: Expr, Children: []*ScriptNode{{Type: FuncName, Func: fn}}}
			for i := 0; i < fn.Arity; i++ {
				if i == insertAt {
					expr.Children = append(expr.Children, node)
				} else {
					expr.Children = append(expr.Children, editor.randomNode())
				}
			}
			return expr
		}
	}
}

func (editor *ScriptEditor) MutateScript(script string) string {
	return "FIXME MUTATE WHATEVER"
}

func (editor *ScriptEditor) SpliceScripts(scriptA, scriptB string) string {
	return "FIXME SPLICE WHATEVERS"
}

func (editor *ScriptEditor) recursiveFormat(node *ScriptNode, indentLevel int) string {
	switch node.Type {
	case Expr:
		subIndentLevel := indentLevel + len(node.Children[0].Func.Name) + 2
		switch node.Children[0].Func.Arity {
		case 0:
			return fmt.Sprintf("(%s)", node.Children[0].Func.Name)
		case 1:
			return fmt.Sprintf("(%s %s)",	node.Children[0].Func.Name,
													editor.recursiveFormat(node.Children[1], subIndentLevel))
		case 2:
			return fmt.Sprintf("(%s %s\n%s%s)",	node.Children[0].Func.Name,
													editor.recursiveFormat(node.Children[1], subIndentLevel),
													strings.Repeat(" ", subIndentLevel),
													editor.recursiveFormat(node.Children[2], subIndentLevel))
		case 3: // 'if' statements traditionally have two-space indentation in Lisp.
			if node.Children[0].Func.Name == "if" {
				return fmt.Sprintf("(%s %s\n%s%s\n%s%s)",	node.Children[0].Func.Name,
														editor.recursiveFormat(node.Children[1], subIndentLevel),
														strings.Repeat(" ", indentLevel + 2),
														editor.recursiveFormat(node.Children[2], indentLevel + 2),
														strings.Repeat(" ", indentLevel + 2),
														editor.recursiveFormat(node.Children[3], indentLevel + 2))
			} else {
				return fmt.Sprintf("(%s %s\n%s%s\n%s%s)",	node.Children[0].Func.Name,
														editor.recursiveFormat(node.Children[1], subIndentLevel),
														strings.Repeat(" ", subIndentLevel),
														editor.recursiveFormat(node.Children[2], subIndentLevel),
														strings.Repeat(" ", subIndentLevel),
														editor.recursiveFormat(node.Children[3], subIndentLevel))
			}
		}
	case FuncName:
		return node.Func.Name
	case Int:
		return fmt.Sprintf("%d", node.N)
	}
	return "OMG WTF AUGH THIS IS THE WORST"
}

func (editor *ScriptEditor) FormatScript(node *ScriptNode) string {
	return editor.recursiveFormat(node, 0) + "\n"
}
