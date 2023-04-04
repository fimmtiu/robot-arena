package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

const MIN_EXPRS_PER_SCRIPT = 20
const MAX_EXPRS_PER_SCRIPT = 1000
const MUTATIONS_PER_SCRIPT = 2   // should this be random?
const MUTATION_SIZE = 10         // should this be random?
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

func RandomScript(minExprs int) string {
	return FormatScript(RandomTree(minExprs))
}

func RandomTree(minExprs int) *ScriptNode {
	script := makeRandomNode()
	for script.Size() < minExprs {
		script = wrapNode(script)
	}
	return script
}

func makeRandomNode() *ScriptNode {
	if rand.Float32() < INTEGER_PERCENT {
		return &ScriptNode{Type: Int, N: randomInt()}
	} else {
		randFunction := AllFunctions[rand.Intn(len(AllFunctions))]
		node := &ScriptNode{Type: Expr, Children: []*ScriptNode{{Type: FuncName, Func: randFunction}}}
		for i := 0; i < randFunction.Arity; i++ {
			node.Children = append(node.Children, makeRandomNode())
		}
		return node
	}
}

// A curve that gives us numbers between 0 and 50, with more small numbers (0-5) than large ones.
// https://www.desmos.com/calculator/onchb78rot
func randomInt() int {
	return int(math.Floor(0.00005 * math.Pow(rand.Float64() * 100, 3)))
}

// Wraps a node in some other multi-argument expression.
func wrapNode(node *ScriptNode) *ScriptNode {
	for {
		fn := AllFunctions[rand.Intn(len(AllFunctions))]
		if fn.Arity > 0 {
			insertAt := rand.Intn(fn.Arity)
			expr := &ScriptNode{Type: Expr, Children: []*ScriptNode{{Type: FuncName, Func: fn}}}
			for i := 0; i < fn.Arity; i++ {
				if i == insertAt {
					expr.Children = append(expr.Children, node)
				} else {
					expr.Children = append(expr.Children, makeRandomNode())
				}
			}
			return expr
		}
	}
}

func MutateScript(script string) string {
	tree := ParseScript(script)
	replacement := RandomTree(MUTATION_SIZE)
	replaceRandomNode(tree, replacement, 0)
	randomlyPruneTree(tree)
	return FormatScript(tree)
}

func SpliceScripts(scriptA, scriptB string) string {
	treeA, treeB := ParseScript(scriptA), ParseScript(scriptB)
	replacement := chooseRandomLocation(treeB).Node

	replaceRandomNode(treeA, replacement, 0)
	randomlyPruneTree(treeA)
	return FormatScript(treeA)
}

// Repeatedly picks a random large-ish branch in the tree and replaces it with something shorter until we get
// below the limit.
func randomlyPruneTree(tree *ScriptNode) {
	for tree.Size() > MAX_EXPRS_PER_SCRIPT {
		replacement := RandomTree(1)
		replaceRandomNode(tree, replacement, replacement.Size())
	}
}

type TreeLocation struct {
	Node *ScriptNode
	Parent *ScriptNode
	Index int
}

func linearizeChildren(tree *ScriptNode) []TreeLocation {
	list := []TreeLocation{}
	for i, node := range tree.Children {
		location := TreeLocation{Node: node, Parent: tree, Index: i}
		list = append(list, location)
		if node.Type == Expr {
			list = append(list, linearizeChildren(node)...)
		}
	}
	return list
}

func chooseRandomLocation(tree *ScriptNode) TreeLocation {
	nodes := linearizeChildren(tree)
	for {
		randomNode := nodes[rand.Intn(len(nodes))]
		if randomNode.Node.Type != FuncName {
			return randomNode
		}
	}
}

func replaceRandomNode(tree, replacement *ScriptNode, minSize int) {
	var randomLocation TreeLocation
	if minSize > 0 {
		for {
			randomLocation = chooseRandomLocation(tree)
			if randomLocation.Node.Size() >= minSize {
				break
			}
		}
	} else {
		randomLocation = chooseRandomLocation(tree)
	}
	randomLocation.Parent.Children[randomLocation.Index] = replacement
}

// When formatting, we want simple expressions (anything where the arguments are all constants or zero-argument
// functions) to be printed on one line for readability's sake.
func simpleExpr(node *ScriptNode) bool {
	if node.Type != Expr {
		return true
	}

	for _, child := range node.Children {
		if child.Type == Expr && child.Children[0].Func.Arity > 0 {
			return false
		}
	}
	return true
}

func recursiveFormat(node *ScriptNode, indentLevel int) string {
	switch node.Type {
	case Expr:
		subIndentLevel := indentLevel + len(node.Children[0].Func.Name) + 2
		switch node.Children[0].Func.Arity {
		case 0:
			return fmt.Sprintf("(%s)", node.Children[0].Func.Name)
		case 1:
			return fmt.Sprintf("(%s %s)",	node.Children[0].Func.Name,
													recursiveFormat(node.Children[1], subIndentLevel))
		case 2:
			if simpleExpr(node) {
				return fmt.Sprintf("(%s %s %s)",	node.Children[0].Func.Name,
														recursiveFormat(node.Children[1], subIndentLevel),
														recursiveFormat(node.Children[2], subIndentLevel))
			} else {
				return fmt.Sprintf("(%s %s\n%s%s)",	node.Children[0].Func.Name,
														recursiveFormat(node.Children[1], subIndentLevel),
														strings.Repeat(" ", subIndentLevel),
														recursiveFormat(node.Children[2], subIndentLevel))
			}
		case 3: // 'if' statements traditionally have two-space indentation in Lisp.
			if node.Children[0].Func.Name == "if" {
				return fmt.Sprintf("(%s %s\n%s%s\n%s%s)",	node.Children[0].Func.Name,
														recursiveFormat(node.Children[1], subIndentLevel),
														strings.Repeat(" ", indentLevel + 2),
														recursiveFormat(node.Children[2], indentLevel + 2),
														strings.Repeat(" ", indentLevel + 2),
														recursiveFormat(node.Children[3], indentLevel + 2))
			} else {
				if simpleExpr(node) {
					return fmt.Sprintf("(%s %s\n%s%s\n%s%s)",	node.Children[0].Func.Name,
															recursiveFormat(node.Children[1], subIndentLevel),
															strings.Repeat(" ", subIndentLevel),
															recursiveFormat(node.Children[2], subIndentLevel),
															strings.Repeat(" ", subIndentLevel),
															recursiveFormat(node.Children[3], subIndentLevel))
				} else {
					return fmt.Sprintf("(%s %s %s %s)",	node.Children[0].Func.Name,
															recursiveFormat(node.Children[1], subIndentLevel),
															recursiveFormat(node.Children[2], subIndentLevel),
															recursiveFormat(node.Children[3], subIndentLevel))
				}
			}
		}
	case FuncName:
		return node.Func.Name
	case Int:
		return fmt.Sprintf("%d", node.N)
	}
	return "OMG WTF AUGH THIS IS THE WORST"
}

func FormatScript(node *ScriptNode) string {
	return recursiveFormat(node, 0) + "\n"
}

// The evolution process creates really weird-looking programs with lots of dead code, so let's build a tool
// to make simplified versions of the scripts so that humans can see how they work.

func SimplifyTree(tree *ScriptNode) {
	if tree.Type == Expr {
		switch tree.Children[0].Func.Name {
		case "if":
			// If the 'if' condition is constant, replace the 'if' statement with the corresponding branch.
			isConstant, value := constantValue(tree.Children[1])
			if isConstant {
				branch := tree.Children[2]
				if value == 0 {
					branch = tree.Children[3]
				}
				*tree = *branch
			}
		}
	}

	// Fold constant values
	isConstant, value := constantValue(tree)
	if isConstant {
		*tree = ScriptNode{Type: Int, N: value}
	}

	for i := 1; i < len(tree.Children); i++ {
		isConstant, value := constantValue(tree.Children[i])
		if isConstant {
			*tree.Children[i] = ScriptNode{Type: Int, N: value}
		}
	}
}

func constantValue(node *ScriptNode) (bool, int) {
	if node.Type == Int {
		return true, node.N
	} else if node.Type == Expr {
		switch (node.Children[0].Func.Name) {
		case "and":
			if isConstant, values := constantArguments(node); isConstant {
				if values[0] == 0 {
					return true, 0
				}
				return true, values[1]
			}
		case "or":
			// 'or' is tricky because not all of the arguments have to be constant; we just need one true constant to return.
			for i := 1; i < len(node.Children); i++ {
				isConstant, value := constantValue(node.Children[i])
				if !isConstant {
					return false, -1
				}
				if value != 0 || i == len(node.Children) - 1 {
					return true, value
				}
			}
				case "+":
			if isConstant, values := constantArguments(node); isConstant {
				return true, values[0] + values[1]
			}
		case "-":
			if isConstant, values := constantArguments(node); isConstant {
				return true, values[0] - values[1]
			}
		case "*":
			if isConstant, values := constantArguments(node); isConstant {
				return true, values[0] * values[1]
			}
		case "/":
			if isConstant, values := constantArguments(node); isConstant {
				if values[1] == 0 {
					return true, 0
				}
				return true, values[0] / values[1]
			}
		case "mod":
			if isConstant, values := constantArguments(node); isConstant {
				if values[1] == 0 {
					return true, 0
				}
				return true, values[0] % values[1]
			}
		case "<":
			if isConstant, values := constantArguments(node); isConstant {
				if values[0] < values[1] {
					return true, 1
				}
				return true, 0
			}
		case ">":
			if isConstant, values := constantArguments(node); isConstant {
				if values[0] > values[1] {
					return true, 1
				}
				return true, 0
			}
		case "=":
			if isConstant, values := constantArguments(node); isConstant {
				if values[0] == values[1] {
					return true, 1
				}
				return true, 0
			}
		case "not":
			if isConstant, values := constantArguments(node); isConstant {
				if values[0] == 0 {
					return true, 1
				}
				return true, 0
			}
		}
	}

	return false, -1
}

func constantArguments(node *ScriptNode) (bool, []int) {
	values := make([]int, len(node.Children) - 1)
	for i := 1; i < len(node.Children); i++ {
		isConstant, value := constantValue(node.Children[i])
		if !isConstant {
			return false, nil
		}
		values[i - 1] = value
	}
	return true, values
}
