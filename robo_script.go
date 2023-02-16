package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ResultType uint8
const (
	ResultAction ResultType = iota
	ResultError
	ResultInt
)

// Let's just use crappy pseudo-unions for now. Not memory-efficient, but simple.
type Result struct {
	Type ResultType
	Action Action
	Err error
	Int int
}

var ResultTrue = Result{Type: ResultInt, Int: 1}
var ResultFalse = Result{Type: ResultInt, Int: 0}

type Function func(s *Script, args []*ScriptNode) Result

type NodeType uint8
const (
	Expr NodeType = iota
	Symbol
	Int
)

type ScriptNode struct {
	Children []*ScriptNode
	Type NodeType
	Sym string
	N int
}

type Script struct {
	Code *ScriptNode
	State *GameState
}

// If the script returns a number instead of performing an action, just wait.
func (s *Script) Run() Result {
	result := s.Eval(s.Code)
	if result.Type != ResultAction {
		result.Type = ResultAction
		result.Action = Action{Type: ActionWait}
	}
	return result
}

func ParseScript(code string) *ScriptNode {
	node, _, err := readToken(code)
	if err != nil {
		logger.Fatalf("Parse error! %v", err)
	}
	return node
}

// It's quick! It's dirty! It's a Lisp parser in ~60 lines!
func readToken(code string) (*ScriptNode, string, error) {
	var err error
	code = strings.TrimSpace(code)
	node := ScriptNode{ Children: make([]*ScriptNode, 0) }

	if code[0] == '(' {
		code = code[1:]
		node.Type = Expr
		for {
			var child *ScriptNode
			child, code, err = readToken(code)
			if err != nil {
				return nil, code, err
			} else if child == nil {
				break
			} else if len(code) == 0 {
				return nil, code, fmt.Errorf("Unterminated expression!")
			}
			node.Children = append(node.Children, child)
		}
		if len(node.Children) == 0 {
			return nil, code, fmt.Errorf("Found an empty list!")
		}
		if node.Children[0].Type != Symbol {
			return nil, code, fmt.Errorf("Non-symbol in function position! Type %v", node.Children[0].Type)
		}
		for _, child := range node.Children[1:] {
			if child.Type == Symbol {
				return nil, code, fmt.Errorf("Symbol '%s' passed as function argument!", child.Sym)
			}
		}

	} else if code[0] == ')' {
		return nil, code[1:], nil

	} else if unicode.IsDigit(rune(code[0])) {
		s := string(code[0])
		for i := 1; i < len(code) - 1 && unicode.IsDigit(rune(code[i])); i++ {
			s += string(code[i])
		}
		var err error
		node.Type = Int
		node.N, err = strconv.Atoi(s)
		if err != nil {
			return nil, code, fmt.Errorf("Couldn't convert int string to int: '%s', %v", s, err)
		}
		return &node, code[len(s):], nil

	} else if !unicode.IsSpace(rune(code[0])) {
		s := string(code[0])
		for i := 1; i < len(code) - 1 && !unicode.IsSpace(rune(code[i])); i++ {
			s += string(code[i])
		}
		node.Type = Symbol
		node.Sym = s
		return &node, code[len(s):], nil
	}

	return &node, code, nil
}

func (s *Script) Eval(node *ScriptNode) Result {
	switch node.Type {
	case Int:
		return Result{Type: ResultInt, Int: node.N}
	case Expr:
		sym := node.Children[0].Sym
		function, err := ResolveFunction(sym)
		if err != nil {
			return Result{Type: ResultError, Err: err}
		}
		return function(s, node.Children[1:])
	case Symbol:
		logger.Fatalf("Tried to evaluate a symbol! '%s'", node.Sym)
	}
	return Result{}
}


// Functions

// Can't use map literal syntax here or we get into recursive initialization.
var functionLookupTable = make(map[string]Function)

func InitScript() {
	// Base functionality
	functionLookupTable["+"] = RS_Add
	functionLookupTable["-"] = RS_Subtract
	functionLookupTable["*"] = RS_Multiply
	functionLookupTable["/"] = RS_Divide
	functionLookupTable["mod"] = RS_Modulus
	functionLookupTable["<"] = RS_LessThan
	functionLookupTable[">"] = RS_GreaterThan
	functionLookupTable["="] = RS_Equal
	functionLookupTable["if"] = RS_If
	functionLookupTable["and"] = RS_And
	functionLookupTable["or"] = RS_Or
	functionLookupTable["not"] = RS_Not

	// Actions
	functionLookupTable["go-north"] = RS_GoNorth
}

func ResolveFunction(name string) (Function, error) {
	function, found := functionLookupTable[name]
	if !found {
		return nil, fmt.Errorf("No such function: '%s'", name)
	}
	return function, nil
}

// "Functions" is a poor choice of words, since the functions are responsible for evaluating their own arguments. It's
// more like a language that has only special forms.
func RS_Add(s *Script, args []*ScriptNode) Result {
	assertArity("+", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}
	result1.Int += result2.Int
	return result1
}

func RS_Subtract(s *Script, args []*ScriptNode) Result {
	assertArity("-", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}
	result1.Int -= result2.Int
	return result1
}

func RS_Multiply(s *Script, args []*ScriptNode) Result {
	assertArity("*", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}
	result1.Int *= result2.Int
	return result1
}

func RS_Divide(s *Script, args []*ScriptNode) Result {
	assertArity("/", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}
	result1.Int /= result2.Int
	return result1
}

func RS_Modulus(s *Script, args []*ScriptNode) Result {
	assertArity("mod", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}
	result1.Int %= result2.Int
	return result1
}

func RS_LessThan(s *Script, args []*ScriptNode) Result {
	assertArity("<", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}

	if result1.Int < result2.Int {
		return ResultTrue
	} else {
		return ResultFalse
	}
}

func RS_GreaterThan(s *Script, args []*ScriptNode) Result {
	assertArity(">", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}

	if result1.Int > result2.Int {
		return ResultTrue
	} else {
		return ResultFalse
	}
}

func RS_Equal(s *Script, args []*ScriptNode) Result {
	assertArity("=", 2, args)
	result1 := s.Eval(args[0])
	if result1.Type != ResultInt {
		return result1
	}
	result2 := s.Eval(args[1])
	if result2.Type != ResultInt {
		return result2
	}

	if result1.Int == result2.Int {
		return ResultTrue
	} else {
		return ResultFalse
	}
}

func RS_If(s *Script, args []*ScriptNode) Result {
	assertArity("if", 3, args)

	condition := s.Eval(args[0])
	if condition.Type != ResultInt {
		return condition
	}
	if condition.Int > 0 {
		return s.Eval(args[1])
	} else {
		return s.Eval(args[2])
	}
}

func RS_And(s *Script, args []*ScriptNode) Result {
	assertArity("and", 2, args)
	var condition Result = ResultFalse

	for _, arg := range args {
		condition = s.Eval(arg)
		if condition.Type != ResultInt {
			return condition
		}
		if condition.Int == 0 {
			return ResultFalse
		}
	}

	return condition
}

func RS_Or(s *Script, args []*ScriptNode) Result {
	assertArity("or", 2, args)

	for _, arg := range args {
		condition := s.Eval(arg)
		if condition.Type != ResultInt || condition.Int > 0 {
			return condition
		}
	}

	return ResultFalse
}

func RS_Not(s *Script, args []*ScriptNode) Result {
	assertArity("not", 1, args)

	condition := s.Eval(args[0])
	if condition.Int > 0 {
		return ResultFalse
	} else {
		return ResultTrue
	}
}

func RS_GoNorth(s *Script, args []*ScriptNode) Result {
	dir := relativeToActualDirection(North, s.State.CurrentBot.Team)
	destination := s.State.Arena.DestinationCellAfterMove(s.State.CurrentBot.Position, dir)
	return Result{Type: ResultAction, Action: Action{Type: ActionMove, Actor: s.State.CurrentBot, Target: destination}}
}

// FIXME: We should move all function lookups and arity checks to compile-time.
func assertArity(name string, n int, args []*ScriptNode) {
	if len(args) < n {
		logger.Fatalf("Not enough arguments to %s!", name)
	} else if len(args) > n {
		logger.Fatalf("Too many arguments to %s!", name)
	}
}
