package jsondsl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseEmptyArray(t *testing.T) {
	input := `[]`

	got, err := Parse(input)

	wantErr := false
	want := []Node{&Array{LBrack: 0, RBrack: 1}}

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestParse(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("TestParse(): got diff:\n%s", diff)
	}
}

func TestParseEmptyObject(t *testing.T) {
	input := `{}`

	got, err := Parse(input)

	wantErr := false
	want := []Node{&Object{LBrace: 0, RBrace: 1}}

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestParse(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("TestParse(): got diff:\n%s", diff)
	}
}

func TestParseOperator(t *testing.T) {
	input := `op()`

	got, err := Parse(input)

	wantErr := false
	want := []Node{&Operator{Op: &Ident{NamePos: 0, Name: "op"}, LParen: 2, RParen: 3}}

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestParse(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("TestParse(): got diff:\n%s", diff)
	}
}

func TestParseEverythingArray(t *testing.T) {
	input := `[
		null,
		false,
		true,
		0,
		1.0,
		-1e7+5,
		"\"abc\"",
		[],
		{},
		id,
		add(1,2),
	]`

	got, err := Parse(input)

	wantErr := false
	want := []Node{&Array{
		Elements: []Value{
			&Null{ValuePos: 4},
			&Bool{ValuePos: 12},
			&Bool{ValuePos: 21, Value: true},
			&Number{LitPos: 29, Literal: "0"},
			&Number{LitPos: 34, Literal: "1.0"},
			&Number{LitPos: 41, Literal: "-1e7+5"},
			&String{Quote: 51, QuotedContent: `"\"abc\""`},
			&Array{LBrack: 64, RBrack: 65},
			&Object{LBrace: 70, RBrace: 71},
			&Ident{NamePos: 76, Name: "id"},
			&Operator{
				Op:     &Ident{NamePos: 82, Name: "add"},
				LParen: 85,
				Args: []Value{
					&Number{LitPos: 86, Literal: "1"},
					&Number{LitPos: 88, Literal: "2"},
				},
				RParen: 89,
			},
		},
		RBrack: 93,
	}}

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestParse(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("TestParse(): got diff:\n%s", diff)
	}
}
