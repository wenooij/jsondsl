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
	if diff := cmp.Diff(want, got); diff != "" {
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
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestParse(): got diff:\n%s", diff)
	}
}

func TestParseOperator(t *testing.T) {
	input := `op()`

	got, err := Parse(input)

	wantErr := false
	want := []Node{&Operator{
		Id: &Ident{NamePos: 0, Name: "op"},
		Args: []*OperatorArgs{{
			LParen: 2,
			RParen: 3,
		}},
	}}

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestParse(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(want, got); diff != "" {
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
		lambda(x,x)(),
	]`

	got, err := Parse(input)

	wantErr := false
	want := []Node{&Array{
		LBrack: 0,
		Elements: []ListElem[Value]{
			{Value: &Null{NullPos: 4}, Comma: 8},
			{Value: &Bool{LitPos: 12}, Comma: 17},
			{Value: &Bool{LitPos: 21, Literal: true}, Comma: 25},
			{Value: &Number{LitPos: 29, Literal: "0"}, Comma: 30},
			{Value: &Number{LitPos: 34, Literal: "1.0"}, Comma: 37},
			{Value: &Number{LitPos: 41, Literal: "-1e7+5"}, Comma: 47},
			{Value: &String{Quote: 51, QuotedContent: `"\"abc\""`}, Comma: 60},
			{Value: &Array{LBrack: 64, RBrack: 65}, Comma: 66},
			{Value: &Object{LBrace: 70, RBrace: 71}, Comma: 72},
			{Value: &Operator{Id: &Ident{NamePos: 76, Name: "id"}}, Comma: 78},
			{Value: &Operator{
				Id: &Ident{NamePos: 82, Name: "add"},
				Args: []*OperatorArgs{
					{
						LParen: 85,
						ValueList: []ListElem[Value]{
							{Value: &Number{LitPos: 86, Literal: "1"}, Comma: 87},
							{Value: &Number{LitPos: 88, Literal: "2"}},
						},
						RParen: 89,
					},
				},
			}, Comma: 90},
			{Value: &Operator{
				Id: &Ident{NamePos: 94, Name: "lambda"},
				Args: []*OperatorArgs{
					{
						LParen: 100,
						ValueList: []ListElem[Value]{
							{Value: &Operator{Id: &Ident{NamePos: 101, Name: "x"}}, Comma: 102},
							{Value: &Operator{Id: &Ident{NamePos: 103, Name: "x"}}},
						},
						RParen: 104,
					},
					{
						LParen: 105,
						RParen: 106,
					},
				},
			}, Comma: 107},
		},
		RBrack: 110,
	}}

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestParse(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestParse(): got diff:\n%s", diff)
	}
}
