package main

// TODO: this should NOT exist
// It's too error prone
// And prevents catching errors during parsing

// Instead a list of functions should be imported and assigned as 'static'
// Just the same as importing a different file

func (me *cfile) compileBuiltin(n *node, name string, parameters []*node) *codeblock {
	switch name {
	case libPush:
		me.libReqAdd(HmLibSlice)
		param0 := me.eval(parameters[0])
		p := param0.data()
		if p.isSlice() {
			uses := p.getmember()
			param1 := me.eval(parameters[1])
			if uses.isPointerInC() {
				cb := codeBlockOne(n, "hmlib_slice_push("+param0.pop()+", "+param1.pop()+")")
				cb.prepend(param0.pre)
				cb.prepend(param1.pre)
				return cb
			}
			cb := codeBlockOne(n, "hmlib_slice_push_"+uses.typeSig(me)+"("+param0.pop()+", "+param1.pop()+")")
			cb.prepend(param0.pre)
			cb.prepend(param1.pre)
			return cb
		}
		panic(me.fail(n) + "Argument for push was not a slice: " + p.error())
	case libPop:
		me.libReqAdd(HmLibSlice)
		param := me.eval(parameters[0])
		p := param.data()
		if p.isSlice() {
			uses := p.getmember()
			if uses.isPointerInC() {
				cb := codeBlockOne(n, "hmlib_slice_pop("+param.pop()+")")
				cb.prepend(param.pre)
				return cb
			}
			sig := uses.typeSig(me)
			cb := codeBlockOne(n, "hmlib_slice_pop_"+sig+"("+param.pop()+")")
			cb.prepend(param.pre)
			return cb
		}
		panic(me.fail(n) + "Argument for pop was not a slice: " + p.error())
	case libLength:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenRawString:
			return codeBlockMerge(n, "((int) strlen("+param.pop()+"))", param.pre)
		case TokenString:
			me.libReqAdd(HmLibString)
			return codeBlockMerge(n, "hmlib_string_len("+param.pop()+")", param.pre)
		}
		p := param.data()
		if p.isArray() {
			return codeBlockMerge(n, p.arraySize(), param.pre)
		} else if p.isSlice() {
			me.libReqAdd(HmLibSlice)
			return codeBlockMerge(n, "hmlib_slice_len("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for len() was " + param.string(0))
	case libCapacity:
		param := me.eval(parameters[0])
		switch param.getType() {
		case TokenRawString:
			return codeBlockMerge(n, "((int) strlen("+param.pop()+"))", param.pre)
		case TokenString:
			me.libReqAdd(HmLibString)
			return codeBlockMerge(n, "hmlib_string_cap("+param.pop()+")", param.pre)
		}
		p := param.data()
		if p.isArray() {
			return codeBlockMerge(n, p.arraySize(), param.pre)
		} else if p.isSlice() {
			me.libReqAdd(HmLibSlice)
			return codeBlockMerge(n, "hmlib_slice_cap("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for cap() was " + param.string(0))
	case libSubstring:
		me.libReqAdd(HmLibString)
		str := me.eval(parameters[0])
		start := me.eval(parameters[1])
		end := me.eval(parameters[2])
		cb := codeBlockOne(n, "hmlib_substring("+str.pop()+", "+start.pop()+", "+end.pop()+")")
		cb.prepend(str.pre)
		cb.prepend(start.pre)
		cb.prepend(end.pre)
		return cb
	case libWrite:
		me.libReqAdd(HmLibFiles)
		path := me.eval(parameters[0])
		content := me.eval(parameters[1])
		cb := codeBlockOne(n, "hmlib_write("+path.pop()+", "+content.pop()+")")
		cb.prepend(path.pre)
		cb.prepend(content.pre)
		return cb
	case libCat:
		me.libReqAdd(HmLibFiles)
		param := me.eval(parameters[0])
		cb := codeBlockOne(n, "hmlib_cat("+param.pop()+")")
		cb.prepend(param.pre)
		return cb
	case libSystem:
		me.libReqAdd(HmLibSystem)
		param := me.eval(parameters[0])
		cb := codeBlockOne(n, "hmlib_popen("+param.pop()+")")
		cb.prepend(param.pre)
		return cb
	case libPrintln:
		fallthrough
	case libPrintf:
		if name == libPrintln {
			parameters[0].value += "\\n"
		}
		cb := &codeblock{}
		code := "printf("
		for ix, param := range parameters {
			if ix > 0 {
				code += ", "
			}
			paramx := me.eval(param)
			cb.prepend(paramx.pre)
			code += paramx.pop()
		}
		code += ")"
		cb.current = codeNode(n, code)
		return cb
	case libSprintln:
		fallthrough
	case libSprintf:
		me.libReqAdd(HmLibString)
		if name == libSprintln {
			parameters[0].value += "\\n"
		}
		code := "hmlib_format("
		cb := &codeblock{}
		for ix, param := range parameters {
			if ix > 0 {
				code += ", "
			}
			paramx := me.eval(param)
			cb.prepend(paramx.pre)
			code += paramx.pop()
		}
		code += ")"
		cb.current = codeNode(n, code)
		return cb
	case libFormat:
		fallthrough
	case libEcho:
		me.libReqAdd(HmLibString)
		code := ""
		if name == libEcho {
			code = "printf(\""
		} else {
			code = "hmlib_format(\""
		}
		cb := &codeblock{}
		code2 := ""
		for ix, param := range parameters {
			if ix > 0 {
				code += " "
			}
			code2 += ", "
			paramx := me.eval(param)
			cb.prepend(paramx.pre)
			pop := true
			switch param.data().getRaw() {
			case TokenChar:
				code += "%c"
			case "[]char":
				fallthrough
			case TokenString:
				code += "%s"
			case TokenRawString:
				code += "%s"
			case TokenInt:
				code += "%d"
			case TokenInt8:
				code += "%\" PRId8 \""
			case TokenInt16:
				code += "%\" PRId16 \""
			case TokenInt32:
				code += "%\" PRId32 \""
			case TokenInt64:
				code += "%\" PRId64 \""
			case TokenSizeT:
				code += "%zu"
			case TokenUInt:
				code += "%u"
			case TokenUInt8:
				code += "%\" PRId8 \""
			case TokenUInt16:
				code += "%\" PRId16 \""
			case TokenUInt32:
				code += "%\" PRId32 \""
			case TokenUInt64:
				code += "%\" PRId64 \""
			case TokenFloat:
				code += "%f"
			case TokenFloat32:
				code += "%f"
			case TokenFloat64:
				code += "%f"
			case "bool":
				code += "%s"
				code2 += paramx.pop() + " ? hmlib_string_init(\"true\") : hmlib_string_init(\"false\")"
				pop = false
			default:
				if param.data().enum != nil {
					panic(me.fail(n) + "echo() does not support printing enums: " + param.data().print())
				} else {
					panic(me.fail(n) + "Argument for echo was " + param.string(me.hmfile, 0))
				}
			}
			if pop {
				code2 += paramx.pop()
			}
		}
		if name == libEcho {
			code += "\\n"
		}
		code += "\""
		code2 += ")"
		cb.current = codeNode(n, code+code2)
		return cb
	case libToStr:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		switch param.getType() {
		case "[]char":
			fallthrough
		case TokenString:
			panic(me.fail(n) + "redundant string cast")
		case TokenChar:
			return codeBlockMerge(n, "hmlib_char_to_string("+param.pop()+")", param.pre)
		case TokenInt:
			return codeBlockMerge(n, "hmlib_int_to_string("+param.pop()+")", param.pre)
		case TokenInt8:
			return codeBlockMerge(n, "hmlib_int8_to_string("+param.pop()+")", param.pre)
		case TokenInt16:
			return codeBlockMerge(n, "hmlib_int16_to_string("+param.pop()+")", param.pre)
		case TokenInt32:
			return codeBlockMerge(n, "hmlib_int32_to_string("+param.pop()+")", param.pre)
		case TokenInt64:
			return codeBlockMerge(n, "hmlib_int64_to_string("+param.pop()+")", param.pre)
		case TokenSizeT:
			return codeBlockMerge(n, "hmlib_size_t_to_string("+param.pop()+")", param.pre)
		case TokenUInt:
			return codeBlockMerge(n, "hmlib_uint_to_string("+param.pop()+")", param.pre)
		case TokenUInt8:
			return codeBlockMerge(n, "hmlib_uint8_to_string("+param.pop()+")", param.pre)
		case TokenUInt16:
			return codeBlockMerge(n, "hmlib_uint16_to_string("+param.pop()+")", param.pre)
		case TokenUInt32:
			return codeBlockMerge(n, "hmlib_uint32_to_string("+param.pop()+")", param.pre)
		case TokenUInt64:
			return codeBlockMerge(n, "hmlib_uint64_to_string("+param.pop()+")", param.pre)
		case TokenFloat:
			return codeBlockMerge(n, "hmlib_float_to_string("+param.pop()+")", param.pre)
		case TokenFloat32:
			return codeBlockMerge(n, "hmlib_float32_to_string("+param.pop()+")", param.pre)
		case TokenFloat64:
			return codeBlockMerge(n, "hmlib_float64_to_string("+param.pop()+")", param.pre)
		case "bool":
			return codeBlockMerge(n, "("+param.pop()+" ? hmlib_string_init(\"true\") : hmlib_string_init(\"false\"))", param.pre)
		}
		panic(me.fail(n) + "argument for string cast was " + param.string(0))
	case libToInt:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to int was " + param.string(0))
	case libToInt8:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int8("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to int8 was " + param.string(0))
	case libToInt16:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int16("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to int16 was " + param.string(0))
	case libToInt32:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int32("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to int32 was " + param.string(0))
	case libToInt64:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_int64("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to int64 was " + param.string(0))
	case libToSizeT:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_size_t("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to size_t was " + param.string(0))
	case libToUInt:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to uint was " + param.string(0))
	case libToUInt8:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint8("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to uint8 was " + param.string(0))
	case libToUInt16:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint16("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to uint16 was " + param.string(0))
	case libToUInt32:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint32("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to uint32 was " + param.string(0))
	case libToUInt64:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_uint64("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to uint64 was " + param.string(0))
	case libToFloat:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_float("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to float was " + param.string(0))
	case libToFloat32:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_float32("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to float32 was " + param.string(0))
	case libToFloat64:
		me.libReqAdd(HmLibString)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "hmlib_string_to_float64("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for conversion to float64 was " + param.string(0))
	case libExit:
		param := me.eval(parameters[0])
		if param.getType() == TokenInt {
			return codeBlockMerge(n, "exit("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for exit was " + param.string(0))
	case libChdir:
		me.stdReq.add(CStdUnistd)
		param := me.eval(parameters[0])
		if param.getType() == TokenString {
			return codeBlockMerge(n, "chdir("+param.pop()+")", param.pre)
		}
		panic(me.fail(n) + "argument for chdir was " + param.string(0))
	default:
		return nil
	}
}
