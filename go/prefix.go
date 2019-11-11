package main

import (
	"strconv"
)

func prefixSign(me *parser, op string) *node {
	node := nodeInit(getPrefixName(op))
	me.eat(op)
	right := me.calc(getPrefixPrecedence(op))
	node.push(right)
	node.copyDataOfNode(right)
	return node
}

func prefixGroup(me *parser, op string) *node {
	me.eat("(")
	node := me.calc(0)
	node.attributes["parenthesis"] = "true"
	me.eat(")")
	return node
}

func prefixPrimitive(me *parser, op string) *node {
	t, ok := literals[op]
	if !ok {
		panic(me.fail() + "unknown primitive \"" + op + "\"")
	}
	node := nodeInit(t)
	node.copyData(me.hmfile.typeToVarData(t))
	node.value = me.token.value
	me.eat(op)
	return node
}

func prefixString(me *parser, op string) *node {
	node := nodeInit(TokenString)
	node.copyData(me.hmfile.typeToVarData(TokenString))
	node.value = me.token.value
	me.eat(TokenStringLiteral)
	return node
}

func prefixChar(me *parser, op string) *node {
	node := nodeInit(TokenChar)
	node.copyData(me.hmfile.typeToVarData(TokenChar))
	node.value = me.token.value
	me.eat(TokenCharLiteral)
	return node
}

func prefixNot(me *parser, op string) *node {
	if me.token.is == "!" {
		me.eat("!")
	} else {
		me.eat("not")
	}
	node := nodeInit("not")
	node.copyData(me.hmfile.typeToVarData("bool"))
	node.push(me.calcBool())
	return node
}

func prefixCast(me *parser, op string) *node {
	me.eat(op)
	node := nodeInit("cast")
	node.copyData(me.hmfile.typeToVarData(op))
	calc := me.calc(getPrefixPrecedence(op))
	value := calc.data().full
	if canCastToNumber(value) {
		node.push(calc)
		return node
	}
	panic(me.fail() + "invalid cast \"" + value + "\"")
}

func prefixIdent(me *parser, op string) *node {
	useStack := false
	if me.token.is == "$" {
		me.eat("$")
		useStack = true
	}

	name := me.token.value
	module := me.hmfile

	if _, ok := module.getType(name); ok {
		if _, ok := module.getFunction(name); ok {
			return me.parseFn(module)
		}
		if _, ok := module.getClass(name); ok {
			data := &allocData{}
			data.stack = useStack
			return me.allocClass(module, data)
		}
		if _, ok := module.enums[name]; ok {
			alloc := &allocData{}
			alloc.stack = useStack
			no := me.allocEnum(module)
			no._vdata.merge(alloc)
			return no
		}
		if def, ok := module.defs[name]; ok {
			return me.exprDef(name, def)
		}
		panic(me.fail() + "bad type \"" + name + "\" definition")
	} else if _, ok := module.imports[name]; ok {
		return me.extern()
	}
	v := module.getvar(name)
	if me.peek().is == ":=" {
		if v != nil && v.mutable == false {
			panic(me.fail() + "variable not mutable")
		}
	} else if v == nil {
		panic(me.fail() + "variable out of scope")
	}
	return me.eatvar(module)
}

func prefixArray(me *parser, op string) *node {
	me.eat("[")
	alloc := &allocData{}
	var no *node
	if me.token.is == "]" {
		alloc.slice = true
		no = nodeInit("slice")
	} else {
		size := me.calc(0)
		if size.getType() != TokenInt {
			panic(me.fail() + "array or slice size " + size.string(0) + " is not an integer")
		}
		slice := false
		var capacity *node
		if me.token.is == ":" {
			me.eat(":")
			slice = true
			if me.token.is != "]" {
				capacity = me.calc(0)
				if capacity.getType() != TokenInt {
					panic(me.fail() + "slice capacity " + capacity.string(0) + " is not an integer")
				}
			}
		}
		if slice || size.is != TokenInt {
			alloc.slice = true
			no = nodeInit("slice")
		} else {
			alloc.array = true
			alloc.size, _ = strconv.Atoi(size.value)
			no = nodeInit("array")
		}
		no.push(size)
		if capacity != nil {
			no.push(capacity)
		}
	}
	me.eat("]")
	data := me.declareType(true)
	data.merge(alloc)
	no._vdata = data

	return no
}

func prefixNone(me *parser, op string) *node {
	me.verify("none")
	n := nodeInit("none")
	n._vdata = me.declareType(true)
	return n
}

func prefixMaybe(me *parser, op string) *node {
	me.verify("maybe")
	n := nodeInit("maybe")
	n._vdata = me.declareType(true)
	return n
}
