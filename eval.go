package jsondsl

import (
	"fmt"
)

type OpFunc = func(scope *Scope, args []any) (any, error)

type Evaluator struct {
	scope *Scope
}

func Eval(scope *Scope, v any) (any, error) {
	return (&Evaluator{scope}).Eval(v)
}

func (e *Evaluator) Init() {
	e.scope = &Scope{}
	e.scope.Reset(nil)
	for id, op := range builtinOps {
		e.scope.Bind(id, op)
	}
}

func (e *Evaluator) Eval(v any) (any, error) {
	switch v := v.(type) {
	case nil, bool, float64, string:
		return v, nil
	case Op:
		return e.evalOp(v)
	case []any:
		return e.evalArray(v)
	case map[string]any:
		return e.evalObject(v)
	default:
		return nil, fmt.Errorf("unexpected type %T", v)
	}
}

// evalOp evaluates op as a builtin operation or as one supplied the evaluator.
func (e *Evaluator) evalOp(op Op) (any, error) {
	v, ok := e.scope.Lookup(op.Id)
	if !ok {
		return nil, fmt.Errorf("op is undefined: %v", op.Id)
	}
	switch v := v.(type) {
	case OpFunc:
		if op.Inv == nil {
			return op, nil
		}
		return e.evalInv(v, op.Inv)
	default:
		return v, nil
	}
}

func (e *Evaluator) evalInv(opFn OpFunc, inv *Inv) (any, error) {
	v, err := opFn(e.scope, inv.Args)
	if err != nil {
		return nil, err
	}
	if inv.Next == nil {
		return v, nil
	}
	next, ok := v.(OpFunc)
	if !ok {
		return nil, fmt.Errorf("call of nonfunction type: %T", v)
	}
	return e.evalInv(next, inv.Next)
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

func (e *Evaluator) evalObject(a map[string]any) (map[string]any, error) {
	aCopy := make(map[string]any, len(a))
	for k, v := range a {
		v, err := e.Eval(v)
		if err != nil {
			return nil, fmt.Errorf("%v at object key %q", err, k)
		}
		aCopy[k] = v
	}
	return aCopy, nil
}
