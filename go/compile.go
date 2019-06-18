package main

import (
	"fmt"
	"strings"
)

func makecode(out string, p *program) {
	cf := cfileInit()
	cf.classes = p.classes
	cf.types = p.types
	cf.primitives = p.primitives

	head := ""
	head += "#include <stdio.h>\n"
	head += "#include <stdlib.h>\n"
	head += "#include <stdbool.h>\n"
	head += "\nint main();\n"

	code := ""
	code += "#include \"main.h\"\n"
	code += "\n"

	for _, c := range p.classOrder {
		code += cf.object(p.classes[c])
	}

	for _, f := range p.functionOrder {
		var s string
		if f == "main" {
			s = cf.mainc(p.functions[f])
		} else {
			s = cf.function(f, p.functions[f])
		}
		code += s
	}
	fmt.Println("===")

	create(out+"/main.h", head)
	create(out+"/main.c", code)
}

func (me *cfile) construct(class string) string {
	return fmt.Sprintf("(%s *)malloc(sizeof(%s))", class, class)
}

func fmtptr(ptr string) string {
	if strings.HasSuffix(ptr, "*") {
		return ptr + "*"
	}
	return ptr + " *"
}

func fmtassignspace(s string) string {
	if strings.HasSuffix(s, "*") {
		return s
	}
	return s + " "
}

func (me *cfile) allocarray(n *node) string {
	size := me.eval(n.has[0])
	atype := parseArrayType(n.typed)
	typed := me.typesig(atype)
	code := "(" + fmtptr(typed) + ")malloc((" + size.code + ") * sizeof(" + typed + "))"
	return code
}

func (me *cfile) checkIsClass(typed string) bool {
	_, ok := me.classes[typed]
	return ok
}

func (me *cfile) typesig(typed string) string {
	if me.checkIsClass(typed) {
		return typed + " *"
	} else if typed == "string" {
		return "char *"
	}
	if checkIsArray(typed) {
		atype := parseArrayType(typed)
		return fmtptr(me.typesig(atype))
	}
	return typed
}

func (me *cfile) eval(n *node) *cnode {
	op := n.is
	if op == "=" {
		right := me.eval(n.has[1])
		code := ""
		nodeLeft := n.has[0]
		var left *cnode
		if nodeLeft.is == "variable" {
			name := nodeLeft.value
			typed := right.typed
			isptr := me.checkIsClass(typed) || checkIsArray(typed)
			if _, ok := me.scope.variables[name]; !ok {
				mutable := false
				if nodeLeft.attribute == "mutable" {
					mutable = true
				}
				me.scope.variables[name] = varInit(typed, name, mutable)
				codesig := fmtassignspace(me.typesig(typed))
				if mutable {
					code = codesig
				} else if isptr {
					code += codesig + "const "
				} else {
					code += "const " + codesig
				}
			}
		}
		left = me.eval(nodeLeft)
		code += left.code + " = " + right.code + ";"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "+=" || op == "-=" || op == "*=" || op == "/=" {
		left := me.eval(n.has[0])
		right := me.eval(n.has[1])
		code := left.code + " " + op + " " + right.code + ";"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "new" {
		code := me.construct(n.typed)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "call" {
		code := me.call(n.value, n.has)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "+" || op == "-" || op == "*" || op == "/" {
		code := me.eval(n.has[0]).code
		code += op
		code += me.eval(n.has[1]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "member-variable" {
		root := n.has[0]
		code := n.value
		for {
			if root.is == "root-variable" {
				if checkIsArray(root.typed) {
					code = root.value + code
				} else {
					code = root.value + "->" + code
				}
				break
			} else if root.is == "array-member" {
				index := me.eval(root.has[0])
				code = "[" + index.code + "]" + "->" + code
				root = root.has[1]
			} else if root.is == "member-variable" {
				code = root.value + "->" + code
				root = root.has[0]
			} else {
				panic("missing member variable")
			}
		}
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "variable" {
		name := n.value
		_, ok := me.scope.variables[name]
		if !ok {
			panic("unknown variable " + name)
		}
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "array-member" {
		index := me.eval(n.has[0])
		root := n.has[1]
		code := root.value + "[" + index.code + "]"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "array" {
		code := me.allocarray(n)
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "return" {
		in := me.eval(n.has[0])
		cn := codeNode(n.is, n.value, n.typed, "return "+in.code+";")
		cn.push(in)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "boolexpr" {
		code := me.eval(n.has[0]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "equal" {
		code := me.eval(n.has[0]).code
		code += " == "
		code += me.eval(n.has[1]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == ">" || op == ">=" || op == "<" || op == "<=" {
		code := me.eval(n.has[0]).code
		code += " " + op + " "
		code += me.eval(n.has[1]).code
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "block" {
		return me.block(n)
	}
	if op == "break" {
		cn := codeNode(n.is, n.value, n.typed, "break;")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "continue" {
		cn := codeNode(n.is, n.value, n.typed, "continue;")
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "for" {
		size := len(n.has)
		ix := 0
		code := "while ("
		if size > 1 {
			ix++
			code += me.eval(n.has[0]).code + ")\n"
		} else {
			code += "true)\n"
		}
		code += fmc(me.depth) + "{\n"
		me.depth++
		code += me.eval(n.has[ix]).code
		me.depth--
		code += fmc(me.depth) + "}"
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "if" {
		hsize := len(n.has)
		code := "if (" + me.eval(n.has[0]).code + ")\n"
		code += fmc(me.depth) + "{\n"
		me.depth++
		code += me.eval(n.has[1]).code
		me.depth--
		code += fmc(me.depth) + "}"
		ix := 2
		for ix < hsize && n.has[ix].is == "elif" {
			code += "\n" + fmc(me.depth) + "else if (" + me.eval(n.has[ix].has[0]).code + ")\n" + fmc(me.depth) + "{\n"
			me.depth++
			code += me.eval(n.has[ix].has[1]).code
			me.depth--
			code += fmc(me.depth) + "}"
			ix++
		}
		if ix >= 2 && ix < hsize && n.has[ix].is == "block" {
			code += "\n" + fmc(me.depth) + "else\n" + fmc(me.depth) + "{\n"
			me.depth++
			code += me.eval(n.has[ix]).code
			me.depth--
			code += fmc(me.depth) + "}"
		}
		cn := codeNode(n.is, n.value, n.typed, code)
		fmt.Println(cn.string(0))
		return cn
	}
	if op == "string" {
		cn := codeNode(n.is, n.value, n.typed, "\""+n.value+"\"")
		fmt.Println(cn.string(0))
		return cn
	}
	if _, ok := me.primitives[op]; ok {
		cn := codeNode(n.is, n.value, n.typed, n.value)
		fmt.Println(cn.string(0))
		return cn
	}
	panic("eval unknown operation " + n.string(0))
}

func (me *cfile) free(name string) string {
	return "free(" + name + ");"
}

func (me *cfile) object(class *class) string {
	code := "struct " + class.name + "\n{\n"
	for _, name := range class.variableOrder {
		field := class.variables[name]
		code += fmc(1) + fmtassignspace(me.typesig(field.is)) + field.name + ";\n"
	}
	code += "};\ntypedef struct " + class.name + " " + class.name + ";\n\n"
	return code
}

func (me *cfile) block(n *node) *cnode {
	expressions := n.has
	code := ""
	for _, expr := range expressions {
		c := me.eval(expr)
		code += fmc(me.depth) + c.code + "\n"
	}
	cn := codeNode(n.is, n.value, n.typed, code)
	fmt.Println(cn.string(0))
	return cn
}

func (me *cfile) mainc(fn *function) string {
	args := fn.args
	expressions := fn.expressions
	codeblock := ""
	returns := false
	me.pushScope()
	for _, arg := range args {
		me.scope.variables[arg.name] = arg
	}
	me.depth = 1
	for _, expr := range expressions {
		c := me.eval(expr)
		if c.is == "return" {
			if c.typed != "int" {
				panic("main must return int")
			} else {
				returns = true
			}
		}
		codeblock += fmc(me.depth) + c.code + "\n"
	}
	if !returns {
		codeblock += fmc(me.depth) + "return 0;\n"
	}
	me.popScope()
	code := ""
	code += "int main("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		code += arg.is + " " + arg.name
	}
	code += ")\n{\n"
	code += codeblock
	code += "}\n"
	return code
}

func (me *cfile) function(name string, fn *function) string {
	args := fn.args
	expressions := fn.expressions
	block := ""
	me.pushScope()
	me.depth = 1
	for _, arg := range args {
		me.scope.variables[arg.name] = arg
	}
	for _, expr := range expressions {
		c := me.eval(expr)
		block += fmc(me.depth) + c.code + "\n"
	}
	me.popScope()
	code := ""
	code += fmtassignspace(me.typesig(fn.typed)) + name + "("
	for ix, arg := range args {
		if ix > 0 {
			code += ", "
		}
		code += "const " + arg.is + " " + arg.name
	}
	code += ")\n{\n"
	code += block
	code += "}\n\n"
	return code
}

func (me *cfile) call(name string, parameters []*node) string {
	if name == "echo" {
		param := me.eval(parameters[0])
		if param.typed == "string" {
			return "printf(\"%s\\n\", " + param.code + ");"
		} else if param.typed == "int" {
			return "printf(\"%d\\n\", " + param.code + ");"
		} else if param.typed == "float" {
			return "printf(\"%f\\n\", " + param.code + ");"
		} else if param.typed == "bool" {
			return "printf(\"%s\\n\", " + param.code + " ? \"true\" : \"false\");"
		}
		panic("argument for echo was " + param.string(0))
	}
	code := name + "("
	for ix, parameter := range parameters {
		if ix > 0 {
			code += ", "
		}
		code += me.eval(parameter).code
	}
	code += ")"
	return code
}
