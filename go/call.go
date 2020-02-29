package main

import (
	"fmt"
	"strconv"
)

func (me *parser) pushSigParams(n *node, sig *fnSig) {
	params := make([]*node, 0)
	me.eat("(")
	ix := 0
	for {
		if me.token.is == ")" {
			me.eat(")")
			break
		} else if ix > 0 {
			me.eat(",")
		}
		arg := sig.args[ix]
		param := me.calc(0, arg.data())
		if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
			err := "parameter \"" + param.data().print()
			err += "\" does not match argument[" + strconv.Itoa(ix) + "] \"" + arg.data().print() + "\" of function signature \"" + sig.print() + "\""
			panic(me.fail() + err)
		}
		params = append(params, param)
	}
	for _, param := range params {
		n.push(param)
	}
}

func (me *parser) pushFunctionParams(n *node, params []*node, fn *function) {
	for ix, param := range params {
		if param == nil {
			var arg *funcArg
			if ix < len(fn.args) {
				arg = fn.args[ix]
			} else {
				arg = fn.argVariadic
			}
			d := arg.defaultNode
			if d == nil {
				d = me.defaultValue(arg.data(), fn.getname())
			}
			n.push(d)
		} else {
			n.push(param)
		}
	}
}

func (me *parser) functionParams(name string, pix int, params []*node, fn *function, lazy bool) (*function, []*node) {
	me.eat("(")
	if me.token.is == "line" {
		me.eat("line")
	}
	min := pix
	dict := false
	size := len(fn.args)
	gtypes := make(map[string]*datatype)
	for {
		if me.token.is == ")" {
			break
		} else if pix > min || dict {
			if me.token.is == "line" {
				me.eat("line")
				if me.token.is == ")" {
					break
				}
			} else {
				me.eat(",")
			}
		}
		if me.token.is == "id" && me.peek().is == ":" {
			argname := me.token.value
			me.eat("id")
			me.eat(":")
			aix := fn.argDict[argname]
			if me.token.is == "_" {
				me.eat("_")
				params[aix] = nil
			} else {
				arg := fn.args[aix]
				param := me.calc(0, nil)

				var update map[string]*datatype
				if len(fn.generics) > 0 {
					update = me.hintGeneric(param.data(), arg.data(), fn.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						a := genericsmap(gtypes)
						b := genericsmap(update)
						f := fmt.Sprint("Lazy generic for function '"+fn.getname()+"' is ", a, " but found ", b)
						panic(me.fail() + f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
					err := "parameter \"" + param.data().print()
					err += "\" does not match argument \"" + argname + "\" typed \"" + arg.data().print() + "\" for function \"" + name + "\""
					panic(me.fail() + err)
				}
				params[aix] = param
			}
			dict = true

		} else if dict {
			panic(me.fail() + "regular paramater found after mapped parameter")
		} else {
			var arg *funcArg
			if pix >= size {
				if fn.argVariadic != nil {
					arg = fn.argVariadic
					params = append(params, nil)
				} else {
					panic(me.fail() + "function \"" + name + "\" argument count exceeds parameter count")
				}
			}
			if me.token.is == "_" {
				me.eat("_")
				params[pix] = nil
			} else {
				param := me.calc(0, nil)
				if arg == nil {
					arg = fn.args[pix]
				}

				var update map[string]*datatype
				if len(fn.generics) > 0 {
					update = me.hintGeneric(param.data(), arg.data(), fn.generics)
				}

				if update != nil && len(update) > 0 {
					lazy = true
					good, newtypes := mergeMaps(update, gtypes)
					if !good {
						a := genericsmap(gtypes)
						b := genericsmap(update)
						f := fmt.Sprint("Lazy generic for function '"+fn.getname()+"' is ", a, " but found ", b)
						panic(me.fail() + f)
					}
					gtypes = newtypes

				} else if param.data().notEquals(arg.data()) && !arg.data().isAnyType() {
					err := "Parameter: " + param.data().print()
					err += " does not match expected: " + arg.data().print() + " for function: " + name
					panic(me.fail() + err)
				}
				params[pix] = param
			}
			pix++
		}
	}
	me.eat(")")
	if lazy {
		module := me.hmfile
		glist := make([]*datatype, len(gtypes))
		for k, v := range gtypes {
			i := inList(fn.generics, k)
			glist[i] = v.copy()
		}
		if len(glist) != len(fn.generics) {
			f := fmt.Sprint("Missing generic for function '"+fn.getname()+"'\nImplementation list was ", genericslist(glist))
			panic(me.fail() + f)
		}
		lazy := name + genericslist(glist)
		if implementation, ok := module.functions[lazy]; ok {
			fn = implementation
		} else {
			fn = remapFunctionImpl(lazy, gtypes, fn)
		}
	}
	return fn, params
}

func (me *parser) callClassFunction(module *hmfile, root *node, c *class, fn *function) *node {
	params := make([]*node, len(fn.args))
	params[0] = root
	_, params = me.functionParams(fn.getclsname(), 1, params, fn, false)
	n := nodeInit("call")
	n.fn = fn
	n.copyData(fn.returns)
	me.pushFunctionParams(n, params, fn)
	return n
}

func (me *parser) call(module *hmfile) *node {
	name := me.token.value
	me.eat("id")
	var order []*datatype
	bfn, ok := module.getFunction(name)
	if !ok {
		panic(me.fail() + "Missing function '" + name + "'")
	}
	fn := bfn
	lazy := false
	if bfn.generics != nil {
		if me.token.is == "<" {
			order, _, _ = me.genericHeader()
			name += genericslist(order)
			gfn, ok := module.getFunction(name)
			if ok {
				fn = gfn
			} else {
				mapping := make(map[string]*datatype)
				for i, g := range bfn.generics {
					mapping[g] = order[i]
				}
				fn = remapFunctionImpl(name, mapping, bfn)
			}
		} else {
			lazy = true
		}
	}
	params := make([]*node, len(fn.args))
	fn, params = me.functionParams(name, 0, params, fn, lazy)
	n := nodeInit("call")
	n.fn = fn
	n.copyData(fn.returns)
	me.pushFunctionParams(n, params, fn)
	return n
}

func (me *parser) parseFn(module *hmfile) *node {
	if me.peek().is == "(" || me.peek().is == "<" {
		return me.call(module)
	}

	name := me.token.value
	fn := module.functions[name]
	me.eat("id")
	n := nodeInit("fn-ptr")
	n.copyData(fn.data())
	n.fn = fn

	return n
}
