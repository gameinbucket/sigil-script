package main

type prefixRule struct {
	precedence int
	name       string
	fn         func(*parser, string) *node
}

type infixRule struct {
	precedence int
	name       string
	fn         func(*parser, *node, string) *node
}

var (
	prefixes map[string]prefixRule
	infixes  map[string]infixRule
)

func init() {
	prefixes = map[string]prefixRule{
		"int":    prefixRule{6, "", prefixPrimitive},
		"float":  prefixRule{6, "", prefixPrimitive},
		"bool":   prefixRule{6, "", prefixPrimitive},
		"string": prefixRule{6, "", prefixPrimitive},
		"none":   prefixRule{6, "", prefixNone},
		"maybe":  prefixRule{6, "", prefixMaybe},
		"id":     prefixRule{6, "", prefixIdent},
		"+":      prefixRule{8, "+sign", prefixSign},
		"-":      prefixRule{8, "-sign", prefixSign},
		"!":      prefixRule{8, "not", prefixNot},
		"not":    prefixRule{8, "", prefixNot},
		"[":      prefixRule{9, "", prefixArray},
		"(":      prefixRule{10, "", prefixGroup},
	}

	infixes = map[string]infixRule{
		"and": infixRule{1, "", infixCompare},
		"or":  infixRule{1, "", infixCompare},
		">":   infixRule{2, "", infixCompare},
		">=":  infixRule{2, "", infixCompare},
		"<":   infixRule{2, "", infixCompare},
		"<=":  infixRule{2, "", infixCompare},
		"=":   infixRule{2, "equal", infixCompare},
		"!=":  infixRule{2, "not-equal", infixCompare},
		">>":  infixRule{2, "", infixBits},
		"<<":  infixRule{2, "", infixBits},
		"&":   infixRule{2, "", infixBits},
		"|":   infixRule{2, "", infixBits},
		"^":   infixRule{2, "", infixBits},
		"+":   infixRule{3, "", infixBinary},
		"-":   infixRule{3, "", infixBinary},
		"*":   infixRule{4, "", infixBinary},
		"/":   infixRule{4, "", infixBinary},
	}
}

func getPrefixPrecedence(op string) int {
	if pre, ok := prefixes[op]; ok {
		return pre.precedence
	}
	return 0
}

func getInfixPrecedence(op string) int {
	if inf, ok := infixes[op]; ok {
		return inf.precedence
	}
	return 0
}

func getPrefixName(op string) string {
	if pre, ok := prefixes[op]; ok {
		if pre.name == "" {
			return op
		}
		return pre.name
	}
	return op
}

func getInfixName(op string) string {
	if inf, ok := infixes[op]; ok {
		if inf.name == "" {
			return op
		}
		return inf.name
	}
	return op
}

func (me *parser) infixOp() string {
	op := me.token.is
	if op == ">" {
		if me.peek().is == ">" {
			me.eat(">")
			op = ">>"
			me.replace(">", op)
		}
	}
	return op
}