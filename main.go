package main

import (
	"fmt"
	"log"
	"unicode"
)

type Token struct {
	tokenType string
	value     string
}

type Node struct {
	nodeType string
	value    string
	name     string
	body     []Node
	params   []Node
	_context *[]NodeNew
}

type NodeNew struct {
	nodeType   string
	value      string
	name       string
	callee     *NodeNew
	expression *NodeNew
	body       *[]NodeNew
	arguments  *[]NodeNew
}

type Visitor struct {
	enter func(node *Node, parent *Node)
	exit  func(node *Node, parent *Node)
}

var visitors = make(map[string]Visitor, 10)

var astNewBody = make([]NodeNew, 0)
var astNew = NodeNew{
	nodeType: "Program",
	body:     &astNewBody,
}

var codeGeneString = make([]string, 0)

func tokenizer(input string) (tokens []Token) {
	var current int = 0
	tokens = make([]Token, 0)
	for current < len(input) {
		ch := input[current]
		if ch == '(' {
			tokens = append(tokens, Token{
				tokenType: "paren",
				value:     "(",
			})
			current++
			continue
		}

		if ch == ')' {
			tokens = append(tokens, Token{
				tokenType: "paren",
				value:     ")",
			})
			current++
			continue
		}

		if unicode.IsSpace(rune(ch)) {
			current++
			continue
		}

		if unicode.IsDigit(rune(ch)) {
			value := ""
			for unicode.IsDigit(rune(ch)) {
				value += string(ch)
				current++
				ch = input[current]
			}
			tokens = append(tokens, Token{
				tokenType: "number",
				value:     value,
			})
			continue
		}

		if ch == '"' {
			value := ""
			current++
			ch = input[current]
			for ch != '"' {
				value += string(ch)
				current++
				ch = input[current]
			}
			tokens = append(tokens, Token{
				tokenType: "string",
				value:     value,
			})
			current++
			continue

		}

		if unicode.IsLetter(rune(ch)) {
			value := ""
			for unicode.IsLetter(rune(ch)) {
				value += string(ch)
				current++
				ch = input[current]
			}
			tokens = append(tokens, Token{
				tokenType: "name",
				value:     value,
			})
			continue
		}

		log.Fatalf("Unknown error ,exiting!!")
	}
	return tokens
}

func walk(current int, tokens []Token) (currentRet int, node Node) {
	token := tokens[current]

	if token.tokenType == "number" {
		current++
		return current, Node{nodeType: "NumberLiteral", value: token.value}
	}

	if token.tokenType == "string" {
		current++
		return current, Node{nodeType: "StringLiteral", value: token.value}
	}

	if token.tokenType == "paren" && token.value == "(" {
		current++
		token = tokens[current]

		nodeParams := make([]Node, 0)

		node = Node{
			nodeType: "CallExpression",
			value:    "",
			name:     token.value,
			params:   nodeParams,
		}

		current++
		token = tokens[current]

		for token.tokenType != "paren" || (token.tokenType == "paren" && token.value != ")") {
			param := Node{}
			current, param = walk(current, tokens)
			node.params = append(node.params, param)
			token = tokens[current]
		}
		current++
		return current, node
	}
	log.Fatalf("Unknown walk error ,exiting!!")
	return current, node
}

func parser(tokens []Token) (ast Node) {
	var current = 0
	astBody := make([]Node, 0)
	ast = Node{
		nodeType: "Program",
		body:     astBody,
	}
	for current < len(tokens) {
		param := Node{}
		current, param = walk(current, tokens)
		ast.body = append(ast.body, param)
	}
	return ast
}

func traversalNode(node Node, parent Node) {

	method := visitors[node.nodeType]
	if method.enter != nil {
		method.enter(&node, &parent)
	}
	switch node.nodeType {
	case "Program":
		traversalArray(node.body, node)
		break
	case "CallExpression":
		traversalArray(node.params, node)
		break
	case "NumberLiteral":
	case "StringLiteral":
		break
	default:
		log.Fatalf("Unknown walk error ,exiting!!")
	}
}

func traversalArray(nodes []Node, parent Node) {
	for _, node := range nodes {
		traversalNode(node, parent)
	}
}

func traversal(ast Node) {
	traversalNode(ast, Node{})
}

func transformer(ast Node) {

	ast._context = &astNewBody
	visitors["NumberLiteral"] = Visitor{
		enter: func(node *Node, parent *Node) {
			nodeNew := NodeNew{
				nodeType:   "NumberLiteral",
				value:      node.value,
				callee:     &NodeNew{},
				expression: nil,
				body:       nil,
				arguments:  nil,
			}
			*parent._context = append(*parent._context, nodeNew)
		},
		exit: nil,
	}
	visitors["StringLiteral"] = Visitor{
		enter: func(node *Node, parent *Node) {
			nodeNew := NodeNew{
				nodeType:   "StringLiteral",
				value:      node.value,
				callee:     &NodeNew{},
				expression: nil,
				body:       nil,
				arguments:  nil,
			}
			*parent._context = append(*parent._context, nodeNew)
		},
		exit: nil,
	}
	visitors["CallExpression"] = Visitor{
		enter: func(node *Node, parent *Node) {
			nodeNew := NodeNew{
				nodeType:   "CallExpression",
				value:      "",
				callee:     &NodeNew{nodeType: "Identifier", name: node.name},
				expression: nil,
				body:       nil,
				arguments:  &[]NodeNew{},
			}

			node._context = nodeNew.arguments
			if parent.nodeType != "CallExpression" {
				nodeNew2 := NodeNew{
					nodeType:   "ExpressionStatement",
					value:      "",
					callee:     &NodeNew{},
					expression: &nodeNew,
					body:       nil,
					arguments:  nil,
				}
				*parent._context = append(*parent._context, nodeNew2)
			} else {
				*parent._context = append(*parent._context, nodeNew)
			}

		},
		exit: nil,
	}

	traversal(ast)
}

func codeGenerator(astNew NodeNew) (retStr string) {
	switch astNew.nodeType {
	case "Program":
		tmpStr := ""
		for _, node := range astNewBody {
			tmpStr += codeGenerator(node) + "\n"
		}
		return tmpStr
	case "ExpressionStatement":
		return codeGenerator(*astNew.expression) + ";"

	case "CallExpression":
		tmpStr := ""
		tmpStr += codeGenerator(*astNew.callee)

		tmpStr2 := "("
		for _, node := range *astNew.arguments {
			tmpStr2 += codeGenerator(node) + ","
		}
		tmpStr2 = tmpStr2[:len(tmpStr2)-1]
		tmpStr2 += ")"

		return tmpStr + tmpStr2

	case "Identifier":
		return astNew.name

	case "NumberLiteral":
		return astNew.value

	case "StringLiteral":
		return "\"" + astNew.value + "\""

	default:
		log.Fatalf("Unknown error ,exiting!!")

	}
	return ""
}

func main() {
	tokens := tokenizer("(add 2 (sub 3 (test 4 5)))")

	ast := parser(tokens)
	transformer(ast)
	fmt.Println(codeGenerator(astNew))
}

