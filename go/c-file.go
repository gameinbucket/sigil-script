package main

import (
	"strings"
)

type cfile struct {
	guard                 string
	pathLocal             string
	pathGlobal            string
	hmfile                *hmfile
	stdReq                *OrderedSet
	libReq                *OrderedSet
	dependencyReq         *OrderedSet
	structReq             *OrderedSet
	enumReq               *OrderedSet
	headStdIncludeSection strings.Builder
	headLibIncludeSection strings.Builder
	headReqIncludeSection strings.Builder
	headSection           strings.Builder
	headSuffix            strings.Builder
	codeFn                []strings.Builder
	rootScope             *scope
	scope                 *scope
	depth                 int
	functions             map[string]*function
	master                bool
}

func (me *hmfile) cFileInit(guard string) *cfile {
	c := &cfile{}
	c.guard = guard
	c.hmfile = me
	c.rootScope = scopeInit(nil)
	c.scope = c.rootScope
	c.codeFn = make([]strings.Builder, 0)
	c.stdReq = newOrderSet()
	c.libReq = newOrderSet()
	c.dependencyReq = newOrderSet()
	c.structReq = newOrderSet()
	c.enumReq = newOrderSet()
	c.functions = make(map[string]*function)
	return c
}

func (me *cfile) pushScope() {
	sc := scopeInit(me.scope)
	me.scope = sc
}

func (me *cfile) popScope() {
	me.scope = me.scope.root
}

func (me *cfile) getvar(name string) *variable {
	// TODO fix scoping rules

	if alias, ok := me.scope.renaming[name]; ok {
		name = alias
	}

	scope := me.scope
	for {
		if v, ok := scope.variables[name]; ok {
			return v
		}
		if scope.root == nil {
			return nil
		}
		scope = scope.root
	}
}

func (me *cfile) includeLibs() {
	for _, name := range me.stdReq.order {
		me.headStdIncludeSection.WriteString("\n#include <" + name + ".h>")
	}
	for _, name := range me.libReq.order {
		location := me.hmfile.program.hmlibmap[name]
		me.hmfile.program.sources[name] = location
		me.headLibIncludeSection.WriteString("\n#include \"" + name + ".h\"")
	}
	if !me.master {
		me.dependencyReq.delete(me.pathLocal)
		for _, name := range me.dependencyReq.order {
			me.headReqIncludeSection.WriteString("\n#include \"" + name + ".h\"")
		}
	}
}

func (me *cfile) head() string {
	me.includeLibs()
	var head strings.Builder
	head.WriteString("#ifndef " + me.guard + "\n")
	head.WriteString("#define " + me.guard + "\n")
	if me.headStdIncludeSection.Len() != 0 {
		head.WriteString(me.headStdIncludeSection.String())
		head.WriteString("\n")
	}
	if me.headLibIncludeSection.Len() != 0 {
		head.WriteString(me.headLibIncludeSection.String())
		head.WriteString("\n")
	}
	if me.headReqIncludeSection.Len() != 0 {
		head.WriteString(me.headReqIncludeSection.String())
		head.WriteString("\n")
	}
	if me.headSection.Len() != 0 {
		head.WriteString(me.headSection.String())
		head.WriteString("\n")
	}
	head.WriteString(me.headSuffix.String())
	return head.String()
}

func (me *cfile) addHeadExtern(expr string) {
	me.headSection.WriteString(expr)
}

func (me *cfile) addHeadFunc(expr string) {
	me.headSection.WriteString(expr)
}

func (me *cfile) addHeadSubInclude(expr string) {
	me.headSection.WriteString(expr)
}

func (me *cfile) addHeadEnum(expr string) {
	me.headSection.WriteString(expr)
}

func (me *cfile) addHeadStruct(expr string) {
	me.headSection.WriteString(expr)
}

func (me *cfile) addHeadEnumTypeDef(expr string) {
	me.headSection.WriteString(expr)
}

func (me *cfile) addHeadStructTypeDef(expr string) {
	me.headSection.WriteString(expr)
}
