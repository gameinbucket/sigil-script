package main

import (
	"fmt"
	"strings"
)

var keywords = map[string]bool{
	"function": true,
	"return":   true,
	"class":    true,
	"new":      true,
	"true":     true,
	"false":    true,
	"free":     true,
}

func (me *token) string() string {
	if me.value == "" {
		return fmt.Sprintf("{type:%s}", me.is)
	}
	return fmt.Sprintf("{type:%s, value:%s}", me.is, me.value)
}

func simpleToken(is string) *token {
	t := &token{}
	t.is = is
	t.value = ""
	return t
}

func valueToken(is, value string) *token {
	t := &token{}
	t.is = is
	t.value = value
	return t
}

func digit(c byte) bool {
	return strings.IndexByte("0123456789", c) >= 0
}

func letter(c byte) bool {
	return strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", c) >= 0
}
func (me *tokenizer) forSpace() {
	stream := me.stream
	for !stream.eof() {
		c := stream.peek()
		if c != ' ' || c == '\t' {
			break
		}
		stream.next()
	}
}

func (me *tokenizer) forNumber() (string, string) {
	stream := me.stream
	typed := "int"
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.peek()
		if c == '.' {
			typed = "float"
			value.WriteByte(c)
			stream.next()
			if !digit(stream.peek()) {
				panic("digit must follow after dot. " + stream.fail())
			}
			continue
		}
		if digit(c) {
			value.WriteByte(c)
			stream.next()
			continue
		}
		break
	}
	return typed, value.String()
}

func (me *tokenizer) forWord() string {
	stream := me.stream
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.peek()
		if !letter(c) {
			break
		}
		value.WriteByte(c)
		stream.next()
	}
	return value.String()
}

func (me *tokenizer) forString() string {
	stream := me.stream
	stream.next()
	value := &strings.Builder{}
	for !stream.eof() {
		c := stream.next()
		if c == '"' {
			break
		}
		value.WriteByte(c)
	}
	return value.String()
}

func tokenize(stream *stream) []*token {
	me := tokenizer{}
	me.stream = stream
	tokens := make([]*token, 0)
	size := len(stream.data)
	for stream.pos < size {
		me.forSpace()
		if stream.pos == size {
			break
		}
		typed, number := me.forNumber()
		if number != "" {
			token := valueToken(typed, number)
			tokens = append(tokens, token)
			continue
		}
		word := me.forWord()
		if word != "" {
			var token *token
			if _, ok := keywords[word]; ok {
				if word == "true" || word == "false" {
					token = valueToken("bool", word)
				} else {
					token = simpleToken(word)
				}
			} else {
				token = valueToken("id", word)
			}
			tokens = append(tokens, token)
			continue
		}
		c := stream.peek()
		if c == '"' {
			value := me.forString()
			token := valueToken("string", value)
			tokens = append(tokens, token)
			continue
		}
		if strings.IndexByte("+-*/()=.:[]", c) >= 0 {
			stream.next()
			token := simpleToken(string(c))
			tokens = append(tokens, token)
			continue
		}
		if c == '\n' {
			stream.next()
			token := simpleToken("line")
			tokens = append(tokens, token)
			continue
		}
		panic("unknown token " + stream.fail())
	}
	token := simpleToken("eof")
	tokens = append(tokens, token)
	return tokens
}