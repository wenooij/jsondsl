package jsondsl

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wenooij/bufiog"
)

func TestTokenize(t *testing.T) {
	s := `reduce([1,2,3], sum)`

	tz := &Tokenizer{}
	sc := bufio.NewScanner(strings.NewReader(s))
	sc.Split(tz.SplitFunc)
	r := &tokenReader{
		t:  tz,
		sc: sc,
	}
	br := bufiog.NewReader(r)

	got := []tokenPos{}
	var err error
	for {
		e, err1 := br.ReadElem()
		if err1 != nil {
			if err1 != io.EOF {
				err = err1
			}
			break
		}
		got = append(got, e)
	}

	wantErr := false
	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("Tokenize(): got err: %v, want err: %v", err, wantErr)
	}

	want := []tokenPos{{
		Token: TokenIdent,
		Text:  "reduce",
		Pos:   0,
	}, {
		Token: TokenLParen,
		Text:  "(",
		Pos:   6,
	}, {
		Token: TokenLBrack,
		Text:  "[",
		Pos:   7,
	}, {
		Token: TokenNumber,
		Text:  "1",
		Pos:   8,
	}, {
		Token: TokenComma,
		Text:  ",",
		Pos:   9,
	}, {
		Token: TokenNumber,
		Text:  "2",
		Pos:   10,
	}, {
		Token: TokenComma,
		Text:  ",",
		Pos:   11,
	}, {
		Token: TokenNumber,
		Text:  "3",
		Pos:   12,
	}, {
		Token: TokenRBrack,
		Text:  "]",
		Pos:   13,
	}, {
		Token: TokenComma,
		Text:  ",",
		Pos:   14,
	}, {
		Token: TokenIdent,
		Text:  "sum",
		Pos:   16,
	}, {
		Token: TokenRParen,
		Text:  ")",
		Pos:   19,
	}}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Tokenize(): got diff:\n%s", diff)
	}
}
