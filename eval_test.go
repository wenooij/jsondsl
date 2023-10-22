package jsondsl

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEvalLiteralLambda(t *testing.T) {
	src := `lambda(4)()`

	d := &Decoder{}
	d.Reset(strings.NewReader(src))

	input, err := d.Decode()
	if err != nil {
		t.Fatalf("TestEval(): failed to parse input: %v", err)
	}
	e := &Evaluator{}
	e.Init()

	got, err := e.Eval(input)
	if err != nil {
		t.Fatalf("TestEval(): failed to evaluate input: %v", err)
	}

	wantErr := false
	want := float64(4)

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestEval(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestEval(): got diff:\n%s", diff)
	}
}

func TestEvalLambda(t *testing.T) {
	src := `lambda(x,x)(4)`

	d := &Decoder{}
	d.Reset(strings.NewReader(src))

	input, err := d.Decode()
	if err != nil {
		t.Fatalf("TestEval(): failed to parse input: %v", err)
	}
	e := &Evaluator{}
	e.Init()

	got, err := e.Eval(input)
	if err != nil {
		t.Fatalf("TestEval(): failed to evaluate input: %v", err)
	}

	wantErr := false
	want := float64(4)

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestEval(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestEval(): got diff:\n%s", diff)
	}
}

func TestEvalOpMapKey(t *testing.T) {
	src := `bind(getkey, lambda("key"))
	{
		getkey(): "value"
	}`

	d := &Decoder{}
	d.Reset(strings.NewReader(src))

	input, err := d.Decode()
	if err != nil {
		t.Fatalf("TestEval(): failed to parse input: %v", err)
	}
	e := &Evaluator{}
	e.Init()

	e.Eval(input)

	input, err = d.Decode()
	if err != nil {
		t.Fatalf("TestEval(): failed to parse input: %v", err)
	}

	got, err := e.Eval(input)
	if err != nil {
		t.Fatalf("TestEval(): failed to evaluate input: %v", err)
	}

	wantErr := false
	want := map[any]any{"key": "value"}

	gotErr := err != nil
	if gotErr != wantErr {
		t.Fatalf("TestEval(): got err = %v, want err = %v", err, wantErr)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestEval(): got diff:\n%s", diff)
	}
}
