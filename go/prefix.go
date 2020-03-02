package main

import (
	"strconv"
)

func prefixSign(me *parser, op string) (*node, *parseError) {
	node := nodeInit(getPrefixName(op))
	me.eat(op)
	right, er := me.calc(getPrefixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	node.push(right)
	node.copyDataOfNode(right)
	return node, nil
}

func prefixGroup(me *parser, op string) (*node, *parseError) {
	me.eat("(")
	node, er := me.calc(0, nil)
	if er != nil {
		return nil, er
	}
	node.attributes["parenthesis"] = "true"
	me.eat(")")
	return node, nil
}

func prefixPrimitive(me *parser, op string) (*node, *parseError) {
	t, ok := literals[op]
	if !ok {
		return nil, err(me, "unknown primitive \""+op+"\"")
	}
	node := nodeInit(t)
	d, er := getdatatype(me.hmfile, t)
	if er != nil {
		return nil, er
	}
	node.copyData(d)
	node.value = me.token.value
	me.eat(op)
	return node, nil
}

func prefixString(me *parser, op string) (*node, *parseError) {
	node := nodeInit(TokenString)
	d, er := getdatatype(me.hmfile, TokenString)
	if er != nil {
		return nil, er
	}
	node.copyData(d)
	node.value = me.token.value
	me.eat(TokenStringLiteral)
	return node, nil
}

func prefixChar(me *parser, op string) (*node, *parseError) {
	node := nodeInit(TokenChar)
	d, er := getdatatype(me.hmfile, TokenChar)
	if er != nil {
		return nil, er
	}
	node.copyData(d)
	node.value = me.token.value
	me.eat(TokenCharLiteral)
	return node, nil
}

func prefixNot(me *parser, op string) (*node, *parseError) {
	if me.token.is == "!" {
		me.eat("!")
	} else {
		me.eat("not")
	}
	node := nodeInit("not")
	newdata, er := getdatatype(me.hmfile, "bool")
	if er != nil {
		return nil, er
	}
	node.copyData(newdata)
	b, er := me.calcBool()
	if er != nil {
		return nil, er
	}
	node.push(b)
	return node, nil
}

func prefixCast(me *parser, op string) (*node, *parseError) {
	me.eat(op)
	node := nodeInit("cast")
	newdata, er := getdatatype(me.hmfile, op)
	if er != nil {
		return nil, er
	}
	node.copyData(newdata)
	calc, er := me.calc(getPrefixPrecedence(op), nil)
	if er != nil {
		return nil, er
	}
	data := calc.data()
	if data.canCastToNumber() {
		node.push(calc)
		return node, nil
	}
	return nil, err(me, "invalid cast \""+data.print()+"\"")
}

func prefixIdent(me *parser, op string) (*node, *parseError) {
	useStack := false
	if me.token.is == "$" {
		me.eat("$")
		useStack = true
	}

	name := me.token.value
	module := me.hmfile

	if _, ok := module.imports[name]; ok && me.peek().is == "." {
		return me.extern()
	}

	if _, ok := module.getType(name); ok {
		if _, ok := module.getFunction(name); ok {
			return me.parseFn(module)
		}
		if _, ok := module.getClass(name); ok {
			hint := &allocHint{}
			hint.stack = useStack
			return me.allocClass(module, hint)
		}
		if _, ok := module.enums[name]; ok {
			hint := &allocHint{}
			hint.stack = useStack
			return me.allocEnum(module, hint)
		}
		if def, ok := module.defs[name]; ok {
			return me.exprDef(name, def)
		}
		return nil, err(me, "Bad type \""+name+"\" definition.")
	}

	v := module.getvar(name)
	if me.peek().is == ":=" {
		if v != nil && v.mutable == false {
			return nil, err(me, "Variable: "+v.name+" is not mutable.")
		}
	} else if v == nil {
		return nil, err(me, "Unknown value: "+name)
	}
	return me.eatvar(module)
}

func prefixArray(me *parser, op string) (*node, *parseError) {
	me.eat("[")
	hint := &allocHint{}
	var no *node
	var size *node
	simple := false
	if me.token.is == "]" {
		hint.slice = true
		no = nodeInit("slice")
		simple = true
	} else if me.token.is == ":" {
		me.eat(":")
		hint.slice = true
		no = nodeInit("slice")
		if me.token.is != "]" {
			capacity, er := me.calc(0, nil)
			if er != nil {
				return nil, er
			}
			if !capacity.data().isInt() {
				return nil, err(me, "slice capacity "+capacity.string(me.hmfile, 0)+" is not an integer")
			}
			defaultSize := nodeInit(TokenInt)
			defaultSize.value = "0"
			newdata, er := getdatatype(me.hmfile, TokenInt)
			if er != nil {
				return nil, er
			}
			defaultSize._vdata = newdata
			no.push(defaultSize)
			no.push(capacity)
		}
	} else {
		var er *parseError
		size, er = me.calc(0, nil)
		if er != nil {
			return nil, er
		}
		if !size.data().isInt() {
			return nil, err(me, "array or slice size "+size.string(me.hmfile, 0)+" is not an integer")
		}
		slice := false
		var capacity *node
		if me.token.is == ":" {
			me.eat(":")
			slice = true
			if me.token.is != "]" {
				capacity, er = me.calc(0, nil)
				if er != nil {
					return nil, er
				}
				if !capacity.data().isInt() {
					return nil, err(me, "slice capacity "+capacity.string(me.hmfile, 0)+" is not an integer")
				}
			}
		}
		if slice || size.is != TokenInt {
			hint.slice = true
			no = nodeInit("slice")
		} else {
			hint.array = true
			hint.size, _ = strconv.Atoi(size.value)
			no = nodeInit("array")
		}
		no.push(size)
		if capacity != nil {
			no.push(capacity)
		}
	}
	me.eat("]")
	data, er := me.declareType()
	if er != nil {
		return nil, er
	}
	if me.token.is == "(" {
		items := nodeInit("items")
		me.eat("(")
		for {
			item, er := me.calc(0, data)
			if er != nil {
				return nil, er
			}
			if item.data().notEquals(data) {
				return nil, err(me, "array member type \""+item.data().print()+"\" does not match array type \""+no.data().getmember().print()+"\"")
			}
			items.push(item)
			if me.token.is == ")" {
				break
			}
			me.eat(",")
		}
		me.eat(")")

		if size != nil {
			sizeint, er := strconv.Atoi(size.value)
			if er != nil || sizeint < len(items.has) {
				return nil, err(me, "defined array size is less than implied size")
			}
		}
		no.push(items)

		if simple {
			no.is = "array"
			hint.array = true
			hint.slice = false
			hint.size = len(items.has)
		}
	}
	data = data.merge(hint)
	no._vdata = data

	return no, nil
}

func prefixNone(me *parser, op string) (*node, *parseError) {
	me.verify("none")
	n := nodeInit("none")
	var er *parseError
	n._vdata, er = me.declareType()
	if er != nil {
		return nil, er
	}
	return n, nil
}

func prefixMaybe(me *parser, op string) (*node, *parseError) {
	me.verify("maybe")
	n := nodeInit("maybe")
	var er *parseError
	n._vdata, er = me.declareType()
	if er != nil {
		return nil, er
	}
	return n, nil
}
