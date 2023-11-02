package jsondsl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/wenooij/bufiog"
)

type Visitor struct {
	*bufiog.Reader[tokenPos]

	visitFn func(Pos, Token, string) error
}

func (v *Visitor) SetVisitor(fn func(Pos, Token, string) error) {
	v.visitFn = fn
}

func callVisitor[E any](fn func(E) error, e E) error {
	if fn != nil {
		return fn(e)
	}
	return nil
}

func callVisitor2[E, E2 any](fn func(E, E2) error, e E, e2 E2) error {
	if fn != nil {
		return fn(e, e2)
	}
	return nil
}

func callVisitor3[E, E2, E3 any](fn func(E, E2, E3) error, e E, e2 E2, e3 E3) error {
	if fn != nil {
		return fn(e, e2, e3)
	}
	return nil
}

func (v *Visitor) Visit(rd io.Reader) error {
	t := &Tokenizer{}
	sc := bufio.NewScanner(rd)
	sc.Split(t.SplitFunc)
	v.Reader = bufiog.NewReaderSize(&tokenReader{
		t:  t,
		sc: sc,
	}, 64)

	for {
		if _, err := v.Peek(1); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := v.visitValue(); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

func (v *Visitor) visitToken(t Token) error {
	e, err := v.ReadElem()
	if err != nil {
		return err
	}
	if e.Token != t {
		return fmt.Errorf("expected token %s (found %s)", t, e.Token)
	}
	if err := callVisitor3(v.visitFn, e.Pos, t, tokenStr[t]); err != nil {
		return err
	}
	return nil
}

func (v *Visitor) visitValue() error {
	es, err := v.Peek(1)
	if err != nil {
		if err == io.EOF {
			return io.ErrUnexpectedEOF
		}
		return err
	}
	switch e := es[0]; e.Token {
	case TokenInvalid:
		return fmt.Errorf("invalid token returned during scan")
	case TokenColon, TokenComma, TokenLParen, TokenRParen, TokenRBrace, TokenRBrack:
		return fmt.Errorf("unexpected token %s at beginning of Value", e.Token)
	case TokenLBrace:
		if err := v.visitObject(); err != nil {
			return err
		}
		return nil
	case TokenLBrack:
		if err := v.visitArray(); err != nil {
			return err
		}
		return nil
	case TokenNull, TokenFalse, TokenTrue, TokenNumber, TokenString:
		v.Discard(1)
		return callVisitor3(v.visitFn, e.Pos, e.Token, e.Text)
	case TokenIdent:
		return v.visitOperator()
	default:
		return fmt.Errorf("unknown token %s returned during scan", e.Token)
	}
}

func (v *Visitor) visitArray() error {
	if err := v.visitToken(TokenLBrack); err != nil {
		return fmt.Errorf("%v at start of array", err)
	}
	if err := visitList[Value](v, TokenRBrack, v.visitValue); err != nil {
		return fmt.Errorf("%v in array", err)
	}
	if err := v.visitToken(TokenRBrack); err != nil {
		return fmt.Errorf("%v at end of array", err)
	}
	return nil
}

func (v *Visitor) visitObject() error {
	if err := v.visitToken(TokenLBrace); err != nil {
		return fmt.Errorf("%v at beginning of object", err)
	}
	if err := visitList[*Member](v, TokenRBrace, v.visitMember); err != nil {
		return fmt.Errorf("%v in object", err)
	}
	if err := v.visitToken(TokenRBrace); err != nil {
		return fmt.Errorf("%v at end of object", err)
	}
	return nil
}

func (v *Visitor) visitIdent() error {
	e, err := v.ReadElem()
	if err != nil {
		return err
	}
	if e.Token != TokenIdent {
		return fmt.Errorf("expected token %s (found %s)", TokenIdent, e.Token)
	}
	callVisitor3(v.visitFn, e.Pos, TokenIdent, e.Text)
	return nil
}

func (v *Visitor) visitMember() error {
	if err := v.visitValue(); err != nil {
		return fmt.Errorf("%v at member key", err)
	}
	if err := v.visitToken(TokenColon); err != nil {
		return fmt.Errorf("%v in object member", err)
	}
	if err := v.visitValue(); err != nil {
		return fmt.Errorf("%v at member Value", err)
	}
	return nil
}

func (v *Visitor) visitOperator() error {
	if err := v.visitIdent(); err != nil {
		return fmt.Errorf("%v at start of operator", err)
	}
	for {
		es, err := v.Peek(1)
		if err != nil && err != io.EOF {
			return err
		}
		if len(es) == 0 || es[0].Token != TokenLParen {
			break
		}
		if err = v.visitOperatorArgs(); err != nil {
			return err
		}
	}
	return nil
}

func (v *Visitor) visitOperatorArgs() error {
	if err := v.visitToken(TokenLParen); err != nil {
		return fmt.Errorf("%v at start of operator arguments", err)
	}
	if err := visitList[Value](v, TokenRParen, v.visitValue); err != nil {
		return fmt.Errorf("%v at operator arguments", err)
	}
	if err := v.visitToken(TokenRParen); err != nil {
		return fmt.Errorf("%v at end of operator", err)
	}
	return nil
}

// visitList visits a generic list of Nodes as seen in the object, array, and operator specs.
// It visits the contents of the list including TokenComma, but does not consume the provided
// delim.
//
// precondition: delim is one of: TokenRBrack, TokenBrace, or TokenRParen.
func visitList[E Node](v *Visitor, delim Token, visitFn func() error) error {
	for done := false; !done; {
		es, err := v.Peek(1)
		if err != nil {
			if err == io.EOF {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		if es[0].Token == delim {
			break
		}
		if err := visitFn(); err != nil {
			return err
		}
		es, err = v.Peek(1)
		if err != nil {
			if err == io.EOF {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		switch es[0].Token {
		case TokenComma:
			v.Discard(1)
			if err := callVisitor3(v.visitFn, es[0].Pos, TokenComma, ","); err != nil {
				return err
			}
		case delim:
			done = true
		default:
		}
	}
	return nil
}
