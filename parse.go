package jsondsl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/wenooij/bufiog"
)

type parser struct {
	*bufiog.Reader[tokenPos]
}

func Parse(src string) ([]Node, error) {
	t := &Tokenizer{}
	sc := bufio.NewScanner(strings.NewReader(src))
	sc.Split(t.SplitFunc)
	p := &parser{
		Reader: bufiog.NewReaderSize(&tokenReader{
			t:  t,
			sc: sc,
		}, 64),
	}

	out := []Node{}
	for {
		if _, err := p.Peek(1); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		val, err := p.parseValue()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		out = append(out, val)
	}
	return out, nil
}

func (p *parser) parseValue() (Value, error) {
	es, err := p.Peek(1)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	switch e := es[0]; e.Token {
	case TokenInvalid:
		return nil, fmt.Errorf("invalid token returned during scan")
	case TokenColon, TokenComma, TokenLParen, TokenRParen, TokenRBrace, TokenRBrack:
		return nil, fmt.Errorf("unexpected token %s at beginning of Value", e.Token)
	case TokenLBrace:
		object, err := p.parseObject()
		if err != nil {
			return nil, err
		}
		return object, nil
	case TokenLBrack:
		array, err := p.parseArray()
		if err != nil {
			return nil, err
		}
		return array, nil
	case TokenNull:
		p.Discard(1)
		return &Null{ValuePos: e.Pos}, nil
	case TokenFalse:
		p.Discard(1)
		return &Bool{ValuePos: e.Pos}, nil
	case TokenTrue:
		p.Discard(1)
		return &Bool{ValuePos: e.Pos, Value: true}, nil
	case TokenNumber:
		p.Discard(1)
		return &Number{LitPos: e.Pos, Literal: e.Text}, nil
	case TokenIdent:
		v, err := p.parseOpOrIdent()
		if err != nil {
			return nil, err
		}
		return v, nil
	case TokenString:
		p.Discard(1)
		return &String{Quote: e.Pos, QuotedContent: e.Text}, nil
	default:
		return nil, fmt.Errorf("unknown token %s returned during scan", e.Token)
	}
}

func (p *parser) parseArray() (*Array, error) {
	lb, err := p.consumeToken(TokenLBrack)
	if err != nil {
		return nil, fmt.Errorf("%v at start of array", err)
	}
	elems, err := parseList(p, TokenRBrack, p.parseValue)
	if err != nil {
		return nil, fmt.Errorf("%v in array", err)
	}
	rb, err := p.consumeToken(TokenRBrack)
	if err != nil {
		return nil, fmt.Errorf("%v at end of array", err)
	}
	return &Array{LBrack: lb, Elements: elems, RBrack: rb}, nil
}

func (p *parser) parseObject() (*Object, error) {
	lb, err := p.consumeToken(TokenLBrace)
	if err != nil {
		return nil, fmt.Errorf("%v at beginning of object", err)
	}
	members, err := parseList(p, TokenRBrace, p.parseMember)
	if err != nil {
		return nil, fmt.Errorf("%v in object", err)
	}
	rb, err := p.consumeToken(TokenRBrace)
	if err != nil {
		return nil, fmt.Errorf("%v at end of object", err)
	}
	return &Object{LBrace: lb, Members: members, RBrace: rb}, nil
}

func (p *parser) parseString() (*String, error) {
	e, err := p.ReadElem()
	if err != nil {
		return nil, err
	}
	if e.Token != TokenString {
		return nil, fmt.Errorf("expected token %s (found %s)", TokenString, e.Token)
	}
	return &String{Quote: e.Pos, QuotedContent: e.Text}, nil
}

func (p *parser) parseIdent() (*Ident, error) {
	e, err := p.ReadElem()
	if err != nil {
		return nil, err
	}
	if e.Token != TokenIdent {
		return nil, fmt.Errorf("expected token %s (found %s)", TokenIdent, e.Token)
	}
	return &Ident{NamePos: e.Pos, Name: e.Text}, nil
}

func (p *parser) consumeToken(t Token) (Pos, error) {
	e, err := p.ReadElem()
	if err != nil {
		return NoPos, err
	}
	if e.Token != t {
		return NoPos, fmt.Errorf("expected token %s (found %s)", t, e.Token)
	}
	return e.Pos, nil
}

func (p *parser) parseMember() (*Member, error) {
	key, err := p.parseString()
	if err != nil {
		return nil, fmt.Errorf("%v at member key", err)
	}
	colon, err := p.consumeToken(TokenColon)
	if err != nil {
		return nil, fmt.Errorf("%v in object member", err)
	}
	value, err := p.parseValue()
	if err != nil {
		return nil, fmt.Errorf("%v at member Value", err)
	}
	return &Member{Key: key, Colon: colon, Value: value}, nil
}

func (p *parser) parseOpOrIdent() (Value, error) {
	id, err := p.parseIdent()
	if err != nil {
		return nil, fmt.Errorf("%v at start of operator", err)
	}
	es, err := p.Peek(1)
	if err != nil {
		if err == io.EOF {
			return id, nil
		}
		return nil, err
	}
	if es[0].Token != TokenLParen {
		return id, nil
	}
	return p.parseOperator(id)
}

func (p *parser) parseOperator(op *Ident) (*Operator, error) {
	lp, err := p.consumeToken(TokenLParen)
	if err != nil {
		return nil, fmt.Errorf("%v at start of operator arguments", err)
	}
	args, err := parseList(p, TokenRParen, p.parseValue)
	if err != nil {
		return nil, fmt.Errorf("%v at operator arguments", err)
	}
	rp, err := p.consumeToken(TokenRParen)
	if err != nil {
		return nil, fmt.Errorf("%v at end of operator", err)
	}
	return &Operator{Op: op, LParen: lp, Args: args, RParen: rp}, nil
}

// parseList parses a generic list of Nodes as seen in the object, array, and operator specs.
// It parses the contents of the list including TokenComma, but does not consume the provided
// delim.
//
// precondition: delim is one of: TokenRBrack, TokenBrace, or TokenRParen.
func parseList[E Node](p *parser, delim Token, parseFn func() (E, error)) ([]E, error) {
	out := []E(nil)

	for first := true; ; first = false {
		es, err := p.Peek(1)
		if err != nil {
			if err == io.EOF {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
		if es[0].Token == delim {
			break
		}
		if !first {
			if es[0].Token != TokenComma {
				return nil, fmt.Errorf("expected token %s (found %s)", TokenComma, es[0].Token)
			}
			p.Discard(1)
			es, err = p.Peek(1)
			if err != nil {
				if err == io.EOF {
					return nil, io.ErrUnexpectedEOF
				}
				return nil, err
			}
		}
		if es[0].Token == delim {
			break
		}
		v, err := parseFn()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

type tokenReader struct {
	t  *Tokenizer
	sc *bufio.Scanner
}

type tokenPos struct {
	Text string
	Token
	Pos
}

func (r *tokenReader) Read(p []tokenPos) (n int, err error) {
	for i := range p {
		if !r.sc.Scan() {
			if err := r.sc.Err(); err != nil {
				return n, err
			}
			return n, io.EOF // The caller might change this to UnexpectedEOF.
		}
		p[i] = tokenPos{Text: r.sc.Text(), Token: r.t.Token(), Pos: r.t.Pos()}
		n++
	}
	return n, nil
}
