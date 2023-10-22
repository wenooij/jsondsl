package jsondsl

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/wenooij/bufiog"
)

type Op struct {
	Id   string
	Args [][]any
}

type Decoder struct {
	*bufiog.Reader[tokenPos]
}

func (d *Decoder) Reset(src io.Reader) {
	t := &Tokenizer{}
	sc := bufio.NewScanner(src)
	sc.Split(t.SplitFunc)
	d.Reader = bufiog.NewReaderSize(&tokenReader{
		t:  t,
		sc: sc,
	}, 64)
}

func (d *Decoder) consumeToken(t Token) (Pos, error) {
	e, err := d.ReadElem()
	if err != nil {
		return NoPos, err
	}
	if e.Token != t {
		return NoPos, fmt.Errorf("expected token %s (found %s)", t, e.Token)
	}
	return e.Pos, nil
}

// Decode a value but returns EOF if no value exists.
func (d *Decoder) Decode() (any, error) {
	if _, err := d.Peek(1); err == io.EOF {
		return nil, io.EOF
	}
	return d.decodeValue()
}

// decodeOptValue deocdes a value otherwise returns UnexpectedEOF.
func (d *Decoder) decodeValue() (any, error) {
	es, err := d.Peek(1)
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
		object, err := d.decodeObject()
		if err != nil {
			return nil, err
		}
		return object, nil
	case TokenLBrack:
		array, err := d.decodeArray()
		if err != nil {
			return nil, err
		}
		return array, nil
	case TokenNull:
		d.Discard(1)
		return nil, nil
	case TokenFalse:
		d.Discard(1)
		return false, nil
	case TokenTrue:
		d.Discard(1)
		return true, nil
	case TokenNumber:
		d.Discard(1)
		v, err := strconv.ParseFloat(e.Text, 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	case TokenIdent:
		v, err := d.decodeOperator()
		if err != nil {
			return nil, err
		}
		return v, nil
	case TokenString:
		s, err := d.decodeString()
		if err != nil {
			return nil, fmt.Errorf("%v at string", err)
		}
		return s, nil
	default:
		return nil, fmt.Errorf("unknown token %s returned during scan", e.Token)
	}
}

func (d *Decoder) decodeArray() ([]any, error) {
	if _, err := d.consumeToken(TokenLBrack); err != nil {
		return nil, fmt.Errorf("%v at start of array", err)
	}
	var elems []any
	if err := decodeList(d, TokenRBrack, func() error {
		v, err := d.decodeValue()
		if err != nil {
			return err
		}
		elems = append(elems, v)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("%v in array", err)
	}
	if _, err := d.consumeToken(TokenRBrack); err != nil {
		return nil, fmt.Errorf("%v at end of array", err)
	}
	return elems, nil
}

func (d *Decoder) decodeObject() (map[any]any, error) {
	if _, err := d.consumeToken(TokenLBrace); err != nil {
		return nil, fmt.Errorf("%v at beginning of object", err)
	}
	var dst map[any]any
	if err := decodeList(d, TokenRBrace, func() error {
		if dst == nil {
			dst = make(map[any]any)
		}
		return d.decodeMember(dst)
	}); err != nil {
		return nil, fmt.Errorf("%v in object", err)
	}
	if _, err := d.consumeToken(TokenRBrace); err != nil {
		return nil, fmt.Errorf("%v at end of object", err)
	}
	return dst, nil
}

func (d *Decoder) decodeString() (string, error) {
	e, err := d.ReadElem()
	if err != nil {
		return "", err
	}
	if e.Token != TokenString {
		return "", fmt.Errorf("expected token %s (found %s)", TokenString, e.Token)
	}
	s, err := strconv.Unquote(e.Text)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (d *Decoder) decodeId() (string, error) {
	e, err := d.ReadElem()
	if err != nil {
		return "", err
	}
	if e.Token != TokenIdent {
		return "", fmt.Errorf("expected token %s (found %s)", TokenIdent, e.Token)
	}
	return e.Text, nil
}

func (d *Decoder) decodeMember(dst map[any]any) error {
	key, err := d.decodeValue()
	if err != nil {
		return fmt.Errorf("%v at member key", err)
	}
	if _, err := d.consumeToken(TokenColon); err != nil {
		return fmt.Errorf("%v in object member", err)
	}
	value, err := d.decodeValue()
	if err != nil {
		return fmt.Errorf("%v at member Value", err)
	}
	dst[key] = value
	return nil
}

func (d *Decoder) decodeOperator() (*Op, error) {
	id, err := d.decodeId()
	if err != nil {
		return nil, fmt.Errorf("%v at start of operator", err)
	}
	var opArgs [][]any
	for {
		es, err := d.Peek(1)
		if err != nil && err != io.EOF {
			return nil, err
		}
		var args []any
		if len(es) == 0 || es[0].Token != TokenLParen {
			break
		}
		if args, err = d.decodeOperatorArgs(); err != nil {
			return nil, err
		}
		opArgs = append(opArgs, args)
	}
	return &Op{Id: id, Args: opArgs}, nil
}

func (d *Decoder) decodeOperatorArgs() ([]any, error) {
	if _, err := d.consumeToken(TokenLParen); err != nil {
		return nil, fmt.Errorf("%v at start of operator arguments", err)
	}
	var args []any
	if err := decodeList(d, TokenRParen, func() error {
		v, err := d.decodeValue()
		if err != nil {
			return err
		}
		args = append(args, v)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("%v at operator arguments", err)
	}
	if _, err := d.consumeToken(TokenRParen); err != nil {
		return nil, fmt.Errorf("%v at end of operator", err)
	}
	return args, nil
}

// decodeList decodes a generic list of Nodes as seen in the object, array, and operator specs.
// It decodes the contents of the list including TokenComma, but does not consume the provided
// delim.
//
// precondition: delim is one of: TokenRBrack, TokenBrace, or TokenRParen.
func decodeList(d *Decoder, delim Token, decodeFn func() error) error {
	for first := true; ; first = false {
		es, err := d.Peek(1)
		if err != nil {
			if err == io.EOF {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		if es[0].Token == delim {
			break
		}
		if !first {
			if es[0].Token != TokenComma {
				return fmt.Errorf("expected token %s (found %s)", TokenComma, es[0].Token)
			}
			d.Discard(1)
			es, err = d.Peek(1)
			if err != nil {
				if err == io.EOF {
					return io.ErrUnexpectedEOF
				}
				return err
			}
		}
		if es[0].Token == delim {
			break
		}
		if err := decodeFn(); err != nil {
			return err
		}
	}
	return nil
}
