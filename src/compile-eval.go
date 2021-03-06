package main

import (
	"strconv"
	"strings"
)

func (me *cfile) hintEval(n *node, hint *datatype) *codeblock {
	op := n.is
	if op == "=" || op == ":=" {
		return me.compileAssign(n)
	}
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" || op == "%=" || op == "&=" || op == "|=" || op == "^=" || op == "<<=" || op == ">>=" {
		code := me.assignmentUpdate(n)
		return codeBlockOne(n, code)
	}
	if op == "new" {
		return me.compileAllocClass(n)
	}
	if op == "enum" {
		return me.compileAllocEnum(n)
	}
	if op == "cast" {
		return me.compileCast(n)
	}
	if op == "concat" {
		size := len(n.has)
		code := ""
		if size == 2 {
			code += "hmlib_concat("
			code += me.eval(n.has[0]).code()
			code += ", "
			code += me.eval(n.has[1]).code()
		} else {
			code += "hmlib_concat_varg(" + strconv.Itoa(size)
			for _, snode := range n.has {
				code += ", " + me.eval(snode).code()
			}
		}
		code += ")"
		return codeBlockOne(n, code)
	}
	if op == "+sign" {
		return me.compilePrefixPos(n)
	}
	if op == "-sign" {
		return me.compilePrefixNeg(n)
	}
	if op == "+" || op == "-" || op == "*" || op == "/" || op == "%" || op == "&" || op == "|" || op == "^" || op == "<<" || op == ">>" {
		return me.compileBinaryOp(n)
	}
	if op == "fn-ptr" {
		return me.compileFnPtr(n, hint)
	}
	if op == "variable" {
		return me.compileVariable(n, hint)
	}
	if op == "root-variable" {
		return me.compileRootVariable(n, hint)
	}
	if op == "array-member" {
		index := me.eval(n.has[0])
		root := me.eval(n.has[1])
		code := root.code() + "[" + index.code() + "]"
		return codeBlockOne(n, code)
	}
	if op == "member-variable" {
		return me.compileMemberVariable(n)
	}
	if op == "union-member-variable" {
		return me.compileUnionMemberVariable(n)
	}
	if op == "call" {
		return me.compileCall(n)
	}
	if op == "array" {
		return me.compileAllocArray(n)
	}
	if op == "slice" {
		return me.compileAllocSlice(n)
	}
	if op == "array-to-slice" {
		return me.compileArrayToSlice(n)
	}
	if op == "return" {
		if len(n.has) > 0 {
			in := me.eval(n.has[0])
			code := "return " + in.pop()
			cb := &codeblock{}
			cb.prepend(in.pre)
			cb.current = codeNode(n, code)
			return cb
		}
		return codeBlockOne(n, "return")
	}
	if op == "boolexpr" {
		code := me.eval(n.has[0]).code()
		return codeBlockOne(n, code)
	}
	if op == "equal" || op == "not-equal" {
		return me.compileEqual(op, n)
	}
	if op == "not" {
		code := me.eval(n.has[0]).code()
		if strings.HasPrefix(code, "!") {
			code = code[1:]
		} else {
			code = "!(" + code + ")"
		}
		return codeBlockOne(n, code)
	}
	if op == ">" || op == ">=" || op == "<" || op == "<=" {
		code := me.eval(n.has[0]).code()
		code += " " + op + " "
		code += me.eval(n.has[1]).code()
		return codeBlockOne(n, code)
	}
	if op == "?" {
		return me.compileTernary(n)
	}
	if op == "and" || op == "or" {
		return me.compileAndOr(n)
	}
	if op == "block" {
		return me.block(n)
	}
	if op == "break" {
		return codeBlockOne(n, "break")
	}
	if op == "continue" {
		return codeBlockOne(n, "continue")
	}
	if op == "goto" {
		return codeBlockOne(n, "goto "+n.value)
	}
	if op == "label" {
		return codeBlockOne(n, n.value+":")
	}
	if op == "pass" {
		return codeBlockOne(n, "")
	}
	if op == "match" {
		return me.compileMatch(n)
	}
	if op == "is" {
		return me.compileIs(n)
	}
	if op == "for" {
		return me.compileFor(op, n)
	}
	if op == "while" {
		return me.compileWhile(op, n)
	}
	if op == "loop" {
		return me.compileLoop(op, n)
	}
	if op == "iterate" {
		return me.compileIterate(op, n)
	}
	if op == "if" {
		return me.compileIf(n)
	}
	if op == "none" {
		return me.compileNone(n)
	}
	if op == "declare" {
		return me.compileLocalDeclare(n)
	}
	if op == "try" {
		return me.compileTry(n)
	}
	if op == "comment" {
		return me.compileComment(n)
	}
	if op == TokenRawString {
		return me.compileRawString(n)
	}
	if op == TokenString {
		return me.compileString(n)
	}
	if op == TokenChar {
		return me.compileChar(n)
	}
	if checkIsPrimitive(op) {
		return codeBlockOne(n, n.value)
	}
	panic(me.fail(n) + "Evaluating unknown operation: " + n.string(me.hmfile, 0))
}
