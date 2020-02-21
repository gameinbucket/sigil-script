package main

type fnSig struct {
	module      *hmfile
	args        []*funcArg
	argVariadic *funcArg
	returns     *datatype
}

func fnSigInit(module *hmfile) *fnSig {
	f := &fnSig{}
	f.module = module
	f.args = make([]*funcArg, 0)
	return f
}

func (me *fnSig) print() string {
	sig := "("
	for ix, arg := range me.args {
		if ix > 0 {
			sig += ", "
		}
		sig += arg.data().getRaw()
	}
	if me.argVariadic != nil {
		if len(me.args) > 0 {
			sig += ", "
		}
		sig += "..." + me.argVariadic.data().getRaw()
	}
	sig += ")"
	if !me.returns.isVoid() {
		sig += " "
		sig += me.returns.getRaw()
	}
	return sig
}

func (me *fnSig) newdatatype() *datatype {
	return getdatatype(nil, me.print())
}

func (me *fnSig) equals(b *fnSig) bool {
	if len(me.args) != len(b.args) {
		return false
	}
	if me.argVariadic != nil || b.argVariadic != nil {
		if me.argVariadic == nil || b.argVariadic == nil || me.argVariadic.data().notEquals(b.argVariadic.data()) {
			return false
		}
	}
	if me.returns.notEquals(b.returns) {
		return false
	}
	for i, pa := range me.args {
		pb := b.args[i]
		if pa.data().notEquals(pb.data()) {
			return false
		}
	}
	return true
}
