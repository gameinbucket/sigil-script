package main

import (
	"strings"
)

func (me *parser) defineEnumImplGeneric(base *enum, impl string, order []string) {

	unionList := make([]*union, len(base.types))
	unionDict := make(map[string]*union)
	for i, v := range base.typesOrder {
		cp := v.copy()
		unionList[i] = cp
		unionDict[cp.name] = cp
	}

	me.hmfile.namespace[impl] = "enum"
	me.hmfile.types[impl] = ""
	me.hmfile.defineOrder = append(me.hmfile.defineOrder, impl+"_enum")

	enumDef := enumInit(base.module, impl, false, unionList, unionDict, nil, nil)
	enumDef.base = base
	base.impls = append(base.impls, enumDef)
	me.hmfile.enums[impl] = enumDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		gmapper[gname] = order[ix]
	}

	for _, un := range unionList {
		for i, typed := range un.types {
			un.types[i] = typeToVarData(me.hmfile, me.genericsReplacer(typed.module, typed.getRaw(), gmapper))
		}
	}
}

func (me *parser) defineClassImplGeneric(base *class, impl string, order []string) {
	memberMap := make(map[string]*variable)
	for k, v := range base.variables {
		memberMap[k] = v.copy()
	}

	module := base.module

	module.namespace[impl] = "type"
	module.types[impl] = ""
	module.defineOrder = append(module.defineOrder, impl+"_type")

	classDef := classInit(module, impl, nil, nil)
	classDef.base = base
	base.impls = append(base.impls, classDef)
	classDef.initMembers(base.variableOrder, memberMap)
	module.classes[impl] = classDef

	gmapper := make(map[string]string)
	for ix, gname := range base.generics {
		gmapper[gname] = order[ix]
	}

	classDef.gmapper = gmapper

	for _, mem := range memberMap {
		mem.update(module, me.genericsReplacer(mem.data().module, mem.data().getRaw(), gmapper))
	}

	for _, fn := range base.functionOrder {
		remapClassFunctionImpl(classDef, fn)
	}
}

func (me *parser) declareGeneric(implementation bool, base hasGenerics) []string {
	me.eat("<")
	gsize := len(base.getGenerics())
	order := make([]string, 0)
	for i := 0; i < gsize; i++ {
		if i != 0 {
			me.eat(",")
		}
		gimpl := me.declareType(implementation)
		order = append(order, gimpl.getRaw())
	}
	me.eat(">")
	return order
}

func (me *parser) declareFn() *varData {
	me.eat("(")
	fn := fnSigInit(me.hmfile)
	if me.token.is != ")" {
		for {
			typed := me.declareType(true)
			fn.args = append(fn.args, fnArgInit(typed.asVariable()))
			if me.token.is == ")" {
				break
			} else if me.token.is == "," {
				me.eat(",")
				continue
			}
			panic(me.fail() + "unexpected token in function pointer")
		}
	}
	me.eat(")")
	if me.token.is != "line" && me.token.is != "," {
		fn.returns = me.declareType(true)
	} else {
		fn.returns = typeToVarData(me.hmfile, "void")
	}

	return fn.data()
}

func (me *parser) declareFnPtr(fn *function) *varData {
	return typeToVarData(me.hmfile, fn.name)
}

func (me *parser) declareType(implementation bool) *varData {
	array := false
	size := ""
	if me.token.is == "[" {
		me.eat("[")
		if me.token.is != "]" {
			sizeNode := me.calc(0)
			if sizeNode.value == "" || !sizeNode.data().dtype.isInt() {
				panic(me.fail() + "array size must be constant integer")
			}
			size = sizeNode.value
		}
		me.eat("]")
		array = true
	}

	module := me.hmfile
	local := ""

	if me.token.is == "(" {
		return me.declareFn()

	} else if me.token.is == "maybe" {
		me.eat("maybe")
		me.eat("<")
		option := me.declareType(implementation)
		me.eat(">")
		local += "maybe<" + option.getRaw() + ">"

	} else if me.token.is == "none" {
		me.eat("none")
		local += "none"
		if me.token.is == "<" {
			me.eat("<")
			option := me.declareType(implementation).getRaw()
			me.eat(">")
			local += "<" + option + ">"
		}
	} else {
		local += me.token.value
		me.wordOrPrimitive()
	}

	if m, ok := me.hmfile.imports[local]; ok {
		me.eat(".")
		module = m
		local = me.token.value
		me.eat("id")
	}

	if _, ok := module.enums[local]; ok && me.token.is == "." {
		me.eat(".")
		local += "." + me.token.value
		me.eat("id")

	} else if fn, ok := module.functions[local]; ok {
		return me.declareFnPtr(fn)
	}

	if me.token.is == "<" {
		if base, ok := module.classes[local]; ok {
			gtypes := me.declareGeneric(implementation, base)
			local += "<" + strings.Join(gtypes, ",") + ">"
			if implementation {
				if _, ok := module.classes[local]; !ok {
					me.defineClassImplGeneric(base, local, gtypes)
				}
			}
		} else if base, ok := module.enums[local]; ok {
			gtypes := me.declareGeneric(implementation, base)
			local += "<" + strings.Join(gtypes, ",") + ">"
			if implementation {
				if _, ok := module.enums[local]; !ok {
					me.defineEnumImplGeneric(base, local, gtypes)
				}
			}
		} else {
			panic(me.fail() + "type \"" + local + "\" does not exist in module \"" + module.name + "\"")
		}
	}

	if array {
		local = "[" + size + "]" + local
	}

	qualified := local
	if module != me.hmfile {
		qualified = module.name + "." + qualified
	}

	return typeToVarData(me.hmfile, qualified)
}

func sizeOfArray(typed string) string {
	i := strings.Index(typed, "]")
	return typed[1:i]
}

func typeOfArrayOrSlice(typed string) (string, string) {
	i := strings.Index(typed, "]")
	size := ""
	if i > 1 {
		size = typed[1:i]
	}
	member := typed[i+1:]
	return size, member
}

func checkIsArrayOrSlice(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '['
}

func checkIsArray(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '[' && typed[1] != ']'
}

func checkIsSlice(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '[' && typed[1] == ']'
}

func checkHasGeneric(typed string) bool {
	return strings.HasSuffix(typed, ">")
}

func checkIsFunction(typed string) bool {
	if len(typed) < 2 {
		return false
	}
	return typed[0] == '('
}

func (me *parser) assignable(n *node) bool {
	return n.is == "variable" || n.is == "member-variable" || n.is == "array-member"
}

func functionSigType(typed string) ([]string, string) {
	end := strings.Index(typed, ")")

	ret := typed[end+1:]
	ret = strings.TrimSpace(ret)
	if ret == "" {
		ret = "void"
	}

	argd := strings.TrimSpace(typed[1:end])
	args := make([]string, 0)

	if argd != "" {
		for _, a := range strings.Split(argd, ",") {
			args = append(args, strings.TrimSpace(a))
		}
	}

	return args, ret
}
