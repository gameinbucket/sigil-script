package main

import "strings"

type class struct {
	module        *hmfile
	name          string
	cname         string
	location      string
	variables     map[string]*variable
	variableOrder []string
	generics      []string
	genericsDict  map[string]int
	gmapper       map[string]string
	functions     map[string]*function
	functionOrder []*function
	base          *class
	impls         []*class
}

func classInit(module *hmfile, name string, generics []string, genericsDict map[string]int) *class {
	c := &class{}
	c.module = module
	c.name = name
	c.location = c.getLocation()
	c.cname = getdatatype(module, name).cname()
	c.generics = generics
	c.genericsDict = genericsDict
	c.functions = make(map[string]*function)
	c.functionOrder = make([]*function, 0)
	if len(generics) > 0 {
		c.impls = make([]*class, 0)
	}
	return c
}

func (me *class) initMembers(variableOrder []string, variables map[string]*variable) {
	me.variableOrder = variableOrder
	me.variables = variables
}

func (me *class) getGenerics() []string {
	return me.generics
}

func (me *class) getLocation() string {
	// path := ""
	name := me.name
	// if strings.Index(name, "<") != -1 {
	// 	path = name[0:strings.Index(name, "<")]
	// } else {
	// 	path = name
	// }
	name = flatten(name)
	name = strings.ReplaceAll(name, "_", "-")
	return name // path + "/" + name
}
