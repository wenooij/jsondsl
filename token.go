package jsondsl

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

//go:generate stringer -type Token -trimprefix Token
type Token int

const (
	TokenInvalid Token = iota //
	TokenColon                // :
	TokenComma                // ,
	TokenLParen               // (
	TokenRParen               // )
	TokenLBrace               // {
	TokenRBrace               // }
	TokenLBrack               // [
	TokenRBrack               // ]
	TokenNull                 // null
	TokenFalse                // false
	TokenTrue                 // true
	TokenNumber               // 123 -1.4e10
	TokenIdent                // abc
	TokenString               // "abc"
)

var byteToken = map[byte]Token{
	':': TokenColon,
	',': TokenComma,
	'{': TokenLBrace,
	'[': TokenLBrack,
	'(': TokenLParen,
	'}': TokenRBrace,
	']': TokenRBrack,
	')': TokenRParen,
}

type Tokenizer struct {
	advance   int
	lastPos   Pos
	lastToken Token
}

func (t *Tokenizer) setToken(pos Pos, token Token) {
	t.lastPos = pos
	t.lastToken = token
}

func (t *Tokenizer) Pos() Pos {
	return t.lastPos
}

func (t *Tokenizer) Token() Token {
	return t.lastToken
}

func (t *Tokenizer) Reset() {
	t.setToken(NoPos, TokenInvalid)
}

func (t *Tokenizer) skipWhitespace(data []byte) (advance int) {
	for {
		r, size := utf8.DecodeRune(data[advance:])
		if !unicode.IsSpace(r) {
			break
		}
		advance += size
	}
	return advance
}

func (t *Tokenizer) SplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	begin := t.skipWhitespace(data)
	advance += begin

	var tok Token
	defer func() {
		if err != nil {
			t.setToken(NoPos, TokenInvalid)
			return
		}
		t.setToken(Pos(t.advance+begin), tok)
		t.advance += advance
	}()

	if len(data[advance:]) == 0 {
		return 0, nil, nil // Try again with a larger buffer if possible.
	}

	if tok = byteToken[data[advance]]; tok != TokenInvalid {
		return advance + 1, data[advance : advance+1], nil
	}

	switch {
	case data[advance] == '-': // Number (negative).
		advance++
		fallthrough
	case '0' <= data[advance] && data[advance] <= '9': // Number
		var dot bool
		var exp bool
		var expSign bool
	loop:
		for {
			if len(data) <= advance {
				return 0, nil, nil // Try again with larger buffer if possible.
			}
			b := data[advance]
			switch {
			case b == '-', b == '+':
				if dot && !exp {
					return advance, nil, fmt.Errorf("invalid character %q after decimal point in numeric literal", b)
				}
				if expSign {
					return advance, nil, fmt.Errorf("invalid character %q in exponent of numeric literal", b)
				}
				expSign = true
			case b == 'e', b == 'E':
				if exp {
					return advance, nil, fmt.Errorf("invalid character %q in exponent of numeric literal", b)
				}
				exp = true
			case b == '.':
				if exp {
					return advance, nil, fmt.Errorf("invalid character %q in exponent of numeric literal", b)
				}
				if dot {
					return advance, nil, fmt.Errorf("invalid character %q after decimal point in numeric literal", b)
				}
				dot = true
			case '0' <= b && b <= '9':
			default:
				break loop
			}
			advance++
		}

		tok = TokenNumber
		return advance, data[begin:advance], nil

	case data[advance] == '"': // String
		bs := data[advance+1:]
		escape := false
		for _, b := range bs {
			advance++
			if b == '\\' {
				escape = true
			} else if b == '"' {
				if !escape {
					break
				}
				escape = false
			}
		}
		tok = TokenString
		return advance + 1, data[begin : advance+1], nil

	case 'A' <= data[advance] && data[advance] <= 'Z' || 'a' <= data[advance] && data[advance] <= 'z' || data[advance] == '_': // Token
		advance++
		for {
			r, size := utf8.DecodeRune(data[advance:])
			if !('A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' || '0' <= r && r <= '9' || r == '_') {
				break
			}
			advance += size
		}

		token = data[begin:advance]
		switch string(token) {
		case "null":
			tok = TokenNull
		case "false":
			tok = TokenFalse
		case "true":
			tok = TokenTrue
		default:
			tok = TokenIdent
		}
		return advance, token, nil
	}

	return 0, nil, fmt.Errorf("unexpected byte %q at start of token", data[advance])
}
