package jsondsl

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEvalOp(t *testing.T) {
	src := `add(2,2)`

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
