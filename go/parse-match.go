package main

func (me *parser) parseIs(left *node, op string, n *node) *node {
	n.copyData(typeToVarData(me.hmfile, "bool"))
	me.eat(op)
	var right *node
	if left.data().checkIsSomeOrNone() {
		invert := false
		if me.token.is == "not" {
			invert = true
			me.eat("not")
		}
		is := ""
		if me.token.is == "some" {
			me.eat("some")
			if invert {
				is = "none"
			} else {
				is = "some"
			}
		} else if me.token.is == "none" {
			me.eat("none")
			if invert {
				is = "some"
			} else {
				is = "none"
			}
		} else {
			panic(me.fail() + "right side of \"is\" was \"" + me.token.is + "\"")
		}
		if is == "some" {
			right = nodeInit("some")
			if me.token.is == "(" {
				if invert {
					panic(me.fail() + "inversion not allowed with value here.")
				}
				me.eat("(")
				temp := me.token.value
				me.eat("id")
				me.eat(")")
				tempd := me.hmfile.varInitFromData(left.data().memberType, temp, false)
				tempv := nodeInit("variable")
				tempv.idata = &idData{}
				tempv.idata.module = me.hmfile
				tempv.idata.name = tempd.name
				tempv.copyData(tempd.data())
				tempv.push(left)
				varnode := &variableNode{tempv, tempd}
				me.hmfile.enumIsStack = append(me.hmfile.enumIsStack, varnode)

				// TODO :: cleanup the above enumIsStack
				tempvv := nodeInit("variable")
				tempvv.idata = &idData{}
				tempvv.idata.module = me.hmfile
				tempvv.idata.name = tempd.name
				tempvv.copyData(tempd.data())
				right.push(tempvv)
				//
			}
		} else if is == "none" {
			right = nodeInit("none")
			if me.token.is == "(" {
				if invert {
					panic(me.fail() + "inversion not allowed with value here.")
				}
				panic(me.fail() + "none type can't have a value here.")
			}
		}
	} else {
		if _, _, ok := left.data().checkIsEnum(); !ok {
			panic(me.fail() + "left side of \"is\" must be enum but was \"" + left.data().print() + "\"")
		}
		if me.token.is == "id" {
			name := me.token.value
			baseEnum, _, _ := left.data().checkIsEnum()
			if un, ok := baseEnum.types[name]; ok {
				prefix := ""
				if me.hmfile != left.data().module {
					prefix = left.data().module.name + "."
				}
				me.eat("id")
				right = nodeInit("match-enum")
				right.copyData(typeToVarData(me.hmfile, prefix+baseEnum.name+"."+un.name))
				if me.token.is == "(" {
					me.eat("(")
					temp := me.token.value
					me.eat("id")
					me.eat(")")
					tempd := me.hmfile.varInitFromData(right.data(), temp, false)
					tempv := nodeInit("variable")
					tempv.idata = &idData{}
					tempv.idata.module = me.hmfile
					tempv.idata.name = tempd.name
					tempv.copyData(tempd.data())
					tempv.push(left)
					varnode := &variableNode{tempv, tempd}
					me.hmfile.enumIsStack = append(me.hmfile.enumIsStack, varnode)

					// TODO :: cleanup the above enumIsStack
					tempvv := nodeInit("variable")
					tempvv.idata = &idData{}
					tempvv.idata.module = me.hmfile
					tempvv.idata.name = tempd.name
					tempvv.copyData(tempd.data())
					right.push(tempvv)
					//
				}
			} else {
				right = me.calc(getInfixPrecedence(op))
			}
		} else if checkIsPrimitive(me.token.is) {
			panic(me.fail() + "can't match on a primitive. did you mean to use an enum implementation?")
		} else {
			panic(me.fail() + "unknown right side of \"is\"")
		}
	}
	n.push(left)
	n.push(right)
	return n
}

func (me *parser) parseMatch() *node {
	depth := me.token.depth
	me.eat("match")
	n := nodeInit("match")

	matching := me.calc(0)
	matchType := matching.data()

	_, un, ok := matchType.checkIsEnum()
	if ok && un != nil {
		panic(me.fail() + "enum \"" + matchType.print() + "\" does not need a match expression.")
	}

	var matchVar *variable
	if matching.is == "variable" {
		matchVar = me.hmfile.getvar(matching.idata.name)
	} else if matching.is == ":=" {
		matchVar = me.hmfile.getvar(matching.has[0].idata.name)
	}

	n.push(matching)

	me.eat("line")
	for {
		if me.token.depth <= depth {
			break
		} else if me.token.is == "id" {
			name := me.token.value
			me.eat("id")
			caseNode := nodeInit(name)
			temp := ""
			if me.token.is == "(" {
				me.eat("(")
				temp = me.token.value
				me.eat("id")
				me.eat(")")
			}
			me.eat("=>")
			n.push(caseNode)
			if temp != "" {
				en, _, ok := matchType.checkIsEnum()
				if !ok {
					panic(me.fail() + "only enums supported for matching")
				}
				tempd := me.hmfile.varInit(en.name+"."+name, temp, false)
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = &idData{}
				tempv.idata.module = me.hmfile
				tempv.idata.name = temp
				tempv.copyData(tempd.data())
				caseNode.push(tempv)
			}
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
			}
			if me.token.is == "line" {
				me.eat("line")
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
			}
		} else if me.token.is == "some" {
			me.eat("some")
			if matchVar != nil {
				if !matchType.checkIsSomeOrNone() {
					panic("type \"" + matchVar.name + "\" is not \"maybe\"")
				}
			}
			temp := ""
			if me.token.is == "(" {
				me.eat("(")
				temp = me.token.value
				me.eat("id")
				me.eat(")")
			}
			me.eat("=>")
			some := nodeInit("some")
			n.push(some)
			if temp != "" {
				tempd := me.hmfile.varInitFromData(matchType.memberType, temp, false)
				me.hmfile.scope.variables[temp] = tempd
				tempv := nodeInit("variable")
				tempv.idata = &idData{}
				tempv.idata.module = me.hmfile
				tempv.idata.name = temp
				tempv.copyData(tempd.data())
				some.push(tempv)
			}
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
			}
			if me.token.is == "line" {
				me.eat("line")
			}
			if temp != "" {
				delete(me.hmfile.scope.variables, temp)
			}
		} else if me.token.is == "none" {
			me.eat("none")
			me.eat("=>")
			n.push(nodeInit("none"))
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
			}
			if me.token.is == "line" {
				me.eat("line")
			}
		} else if me.token.is == "_" {
			me.eat("_")
			me.eat("=>")
			n.push(nodeInit("_"))
			if me.token.is == "line" {
				me.eat("line")
				n.push(me.block())
			} else {
				n.push(me.expression())
			}
			if me.token.is == "line" {
				me.eat("line")
			}
		} else {
			panic(me.fail() + "unknown match expression")
		}
	}

	return n
}
