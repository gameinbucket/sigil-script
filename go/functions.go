package main

type function struct {
	name          string
	clsname       string
	cname         string
	start         *parsepoint
	module        *hmfile
	forClass      *class
	args          []*funcArg
	argDict       map[string]int
	argVariadic   *funcArg
	aliasing      map[string]string
	expressions   []*node
	returns       *datatype
	generics      map[string]int
	genericsOrder []string
	genericsAlias map[string]string
	base          *function
	impls         []*function
	comments      []string
}

func funcInit(module *hmfile, name string, class *class) *function {
	f := &function{}
	f.module = module
	f.name = name
	if class != nil {
		f.clsname = nameOfClassFunc(class.name, name)
	}
	if module != nil {
		if class == nil {
			f.cname = module.funcNameSpace(name)
		} else {
			f.cname = module.funcNameSpace(f.clsname)
		}
	}
	f.args = make([]*funcArg, 0)
	f.argDict = make(map[string]int)
	f.expressions = make([]*node, 0)
	f.aliasing = make(map[string]string)
	f.forClass = class
	return f
}

func (me *function) copy() *function {
	f := &function{}
	f.module = me.module
	f.name = me.name
	f.args = make([]*funcArg, len(me.args))
	for i, a := range me.args {
		f.args[i] = a.copy()
	}
	f.argDict = me.argDict
	f.expressions = make([]*node, len(me.expressions))
	for i, e := range me.expressions {
		f.expressions[i] = e.copy()
	}
	f.returns = me.returns.copy()
	return f
}

func (me *function) getname() string {
	if me.forClass == nil {
		return me.name
	}
	return me.clsname
}

func (me *function) getcname() string {
	return me.cname
}

func (me *function) getclsname() string {
	return me.clsname
}

func (me *function) canonical(current *hmfile) string {
	name := me.getname()
	if me.module != nil {
		return me.module.reference(name)
	}
	return name
}

func (me *function) asSig() *fnSig {
	sig := fnSigInit(me.module)
	for _, arg := range me.args {
		sig.args = append(sig.args, arg)
	}
	sig.argVariadic = me.argVariadic
	sig.returns = me.returns
	return sig
}

func (me *function) data() *datatype {
	return me.asSig().newdatatype()
}

func nameOfClassFunc(cl, fn string) string {
	return cl + "_" + fn
}

func (me *parser) pushFunction(name string, module *hmfile, fn *function) {
	module.functionOrder = append(module.functionOrder, name)
	module.functions[name] = fn
	module.types[name] = "function"
	if me.file != nil {
		me.file.WriteString(fn.string(me.hmfile, 0))
	}
}

func (me *parser) remapFunctionImpl(funcName string, alias map[string]string, original *function) *function {
	module := original.module
	parsing := module.parser
	module.program.pushRemapStack(module.reference(original.name))
	pos := parsing.save()
	parsing.jump(original.start)
	fn := me.defineFunction(funcName, alias, original, nil)
	parsing.jump(pos)
	fn.start = original.start
	parsing.pushFunction(fn.getname(), module, fn)
	module.program.popRemapStack()
	return fn
}

func remapClassFunctionImpl(class *class, original *function) {
	module := class.module
	parsing := module.parser
	funcName := original.name
	module.program.pushRemapStack(module.reference(funcName))
	pos := parsing.save()
	parsing.jump(original.start)
	fn := parsing.defineFunction(funcName, nil, nil, class)
	parsing.jump(pos)
	fn.start = original.start
	class.functions[fn.name] = fn
	class.functionOrder = append(class.functionOrder, fn)
	parsing.pushFunction(fn.getname(), module, fn)
	module.program.popRemapStack()
}

func (me *parser) defineClassFunction() {
	module := me.hmfile
	className := me.token.value
	class := module.classes[className]
	me.eat("id")
	funcName := me.token.value
	globalFuncName := nameOfClassFunc(class.name, funcName)
	me.eat("id")
	if _, ok := module.functions[globalFuncName]; ok {
		panic(me.fail() + "class \"" + className + "\" with function \"" + funcName + "\" is already defined")
	}
	if _, ok := class.variables[funcName]; ok {
		panic(me.fail() + "class \"" + className + "\" with variable \"" + funcName + "\" is already defined")
	}
	fn := me.defineFunction(funcName, nil, nil, class)
	class.functions[funcName] = fn
	class.functionOrder = append(class.functionOrder, fn)
	me.pushFunction(globalFuncName, module, fn)
	for _, implementation := range class.implementations {
		remapClassFunctionImpl(implementation, fn)
	}
}

func (me *parser) defineStaticFunction() {
	module := me.hmfile
	token := me.token
	name := token.value
	if _, ok := module.functions[name]; ok {
		panic(me.fail() + "Function \"" + name + "\" is already defined.")
	}
	if name == "static" {
		panic(me.fail() + "Function \"" + name + "\" is reserved.")
	}
	me.eat("id")
	fn := me.defineFunction(name, nil, nil, nil)
	me.pushFunction(name, module, fn)
}

func (me *parser) defineFunction(name string, alias map[string]string, base *function, self *class) *function {
	module := me.hmfile
	if base != nil {
		module = base.module
	} else if self != nil {
		module = self.module
	}
	fn := funcInit(module, name, self)
	if len(me.hmfile.comments) > 0 {
		fn.comments = me.hmfile.comments
		me.hmfile.comments = make([]string, 0)
	}
	module.pushScope()
	module.scope.fn = fn
	fname := name
	if base != nil {
		fn.base = base
		base.impls = append(base.impls, fn)
		fn.aliasing = alias
	} else if self != nil {
		ref := module.fnArgInit(self.uid(), "self", false)
		fn.argDict["self"] = 0
		fn.args = append(fn.args, ref)
		fn.aliasing = self.gmapper
		fname = self.name + "." + name
	}
	if me.token.is == "<" {
		if self != nil {
			panic(me.fail() + "class functions cannot have additional generics")
		}
		if base != nil {
			panic(me.fail() + "implementation of static function cannot have additional generics")
		}
		order, dict := me.genericHeader()
		me.verify("(")
		fn.generics = dict
		fn.genericsOrder = datatypels(order)
	}
	fn.start = me.save()
	parenthesis := false
	if me.token.is == "(" {
		me.eat("(")
		parenthesis = true
		if me.token.is != ")" {
			for {
				if me.token.is != "id" {
					panic(me.fail() + "unexpected token in function definition")
				}
				argname := me.token.value
				me.eat("id")
				defaultValue := ""
				defaultType := ""
				if me.token.is == ":" {
					me.eat(":")
					op := me.token.is
					if literal, ok := literals[op]; ok {
						defaultValue = me.token.value
						defaultType = literal
						me.eat(op)
					} else {
						panic(me.fail() + "only primitive literals allowed for parameter defaults. was \"" + me.token.is + "\"")
					}
				}
				typed := me.declareType()
				fn.argDict[argname] = len(fn.args)
				fnArg := &funcArg{}
				fnArg.variable = typed.getnamedvariable(argname, false)
				if defaultValue != "" {
					defaultTypeVarData := getdatatype(module, defaultType)
					if typed.notEquals(defaultTypeVarData) {
						panic(me.fail() + "function parameter default type \"" + defaultType + "\" and signature \"" + typed.print() + "\" do not match")
					}
					defaultNode := nodeInit(defaultTypeVarData.getRaw())
					defaultNode.copyData(defaultTypeVarData)
					defaultNode.value = defaultValue
					fnArg.defaultNode = defaultNode
				}
				fn.args = append(fn.args, fnArg)
				if me.token.is == ")" {
					break
				} else {
					me.eat(",")
				}
			}
		}
		me.eat(")")
	} else {
		if self != nil {
			panic(me.fail() + "class function \"" + fname + "\" must include parenthesis")
		}
	}
	if me.token.is != "line" {
		if !parenthesis {
			panic(me.fail() + "function \"" + name + "\" returns a value and must include parenthesis")
		}
		fn.returns = me.declareType()
	} else {
		fn.returns = getdatatype(module, "void")
	}
	me.eat("line")
	for _, arg := range fn.args {
		module.scope.variables[arg.name] = arg.variable
	}
	for {
		for me.token.is == "line" {
			me.eat("line")
			if me.token.is != "line" {
				if me.token.depth != 1 {
					goto fnEnd
				}
				break
			}
		}
		if me.token.depth == 0 {
			goto fnEnd
		}
		expr := me.expression()
		fn.expressions = append(fn.expressions, expr)
	}
fnEnd:
	for _, arg := range fn.args {
		if !arg.used {
			er := me.fail() + "Variable \"" + arg.name + "\" for function \"" + fname + "\" was unused."
			if arg.name == "self" {
				er += " Can this be a static function?"
			}
			panic(er)

		}
	}
	module.popScope()
	return fn
}
