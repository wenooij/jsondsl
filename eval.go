package jsondsl

import (
	"fmt"
	"io"
	"strings"
)

type OpFunc = func(scope *Scope, args []any) (any, error)

type Evaluator struct {
	scope *Scope
}

// Eval evaluates a value returned from a Decoder.
func Eval(scope *Scope, val any) (any, error) {
	return (&Evaluator{scope}).Eval(val)
}

func EvalOpFunc(scope *Scope, val any) (OpFunc, error) {
	v, err := Eval(scope, val)
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	case *Op:
		val, err := scope.Lookup(v.Id)
		if err != nil {
			return nil, err
		}
		if val, ok := val.(OpFunc); ok {
			return val, nil
		}
		return nil, fmt.Errorf("name %q is %s not op", v.Id, TypeName(v))
	default:
		return nil, fmt.Errorf("expected op, found %s", TypeName(v))
	}
}

// EvalSource evaluates all statements in src and returns only
// the value of the last statement.
// All bindings to the provided scope are retained.
func EvalSource(scope *Scope, src string) (any, error) {
	d := &Decoder{}
	d.Reset(strings.NewReader(src))
	e := &Evaluator{scope}
	var res any
	for {
		val, err := d.Decode()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		val, err = e.Eval(val)
		if err != nil {
			return nil, err
		}
		res = val
	}
	return res, nil
}

func (e *Evaluator) Init() {
	e.scope = BuiltinScope()
}

func (e *Evaluator) Eval(v any) (any, error) {
	switch v := v.(type) {
	case nil, bool, float64, string:
		return v, nil
	case *Op:
		return e.evalOp(v)
	case []any:
		return e.evalArray(v)
	case map[any]any:
		return e.evalObject(v)
	default:
		return nil, fmt.Errorf("unexpected type %T", v)
	}
}

// evalOp evaluates op as a builtin operation or as one supplied the evaluator.
func (e *Evaluator) evalOp(op *Op) (any, error) {
	v, err := e.scope.Lookup(op.Id)
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	case OpFunc:
		if len(op.Args) == 0 {
			return v, nil
		}
		return e.evalOpArgs(v, op.Args)
	default:
		if len(op.Args) != 0 {
			return nil, fmt.Errorf("call of nonfunction type: %T", v)
		}
		return v, nil
	}
}

func (e *Evaluator) evalOpArgs(opFn OpFunc, args [][]any) (any, error) {
	v, err := opFn(e.scope, args[0])
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		return v, nil
	}
	next, ok := v.(OpFunc)
	if !ok {
		return nil, fmt.Errorf("call of nonfunction type: %T", v)
	}
	return e.evalOpArgs(next, args[1:])
}

func (e *Evaluator) evalArray(a []any) ([]any, error) {
	aCopy := make([]any, len(a))
	for i, v := range a {
		v, err := e.Eval(v)
		if err != nil {
			return nil, fmt.Errorf("%v at array index %d", err, i)
		}
		aCopy[i] = v
	}
	return aCopy, nil
}

func (e *Evaluator) evalObject(a map[any]any) (map[any]any, error) {
	aCopy := make(map[any]any, len(a))
	for k, v := range a {
		kv, err := e.Eval(k)
		if err != nil {
			return nil, fmt.Errorf("%v at object key %v", err, k)
		}
		// Check whether kv is hashable.
		switch kv.(type) {
		case nil, bool, float64, string:
		default:
			return nil, fmt.Errorf("unhashable type %T at object key %v", kv, k)
		}
		v, err := e.Eval(v)
		if err != nil {
			return nil, fmt.Errorf("%v at object value %v", err, k)
		}
		aCopy[kv] = v
	}
	return aCopy, nil
}
