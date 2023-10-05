package jsondsl

import (
	"fmt"
)

type Scope struct {
	Parent *Scope
	Vars   map[string]any
}

func (s *Scope) Reset(parent *Scope) {
	s.Parent = parent
	s.Vars = make(map[string]any)
}

func (s *Scope) Lookup(id string) (any, bool) {
	if s == nil {
		return nil, false
	}
	if v, ok := s.Vars[id]; ok {
		return v, true
	}
	return s.Parent.Lookup(id)
}

func (s *Scope) LookupVar(id string) (any, bool) {
	if v, ok := s.Lookup(id); ok {
		if _, ok := v.(*OpSig); !ok {
			return v, true
		}
	}
	return nil, false
}

func (s *Scope) LookupOp(id string) (*OpSig, bool) {
	if v, ok := s.Lookup(id); ok {
		if sig, ok := v.(*OpSig); ok {
			return sig, true
		}
	}
	return nil, false
}

func (s *Scope) Bind(id string, val any) (oldVal any, overwrote bool) {
	oldVal, overwrote = s.Vars[id]
	s.Vars[id] = val
	return oldVal, overwrote
}

type OpSig struct {
	NArgs    int
	Variadic bool
	NReturns int
	Func     OpFunc
}

type OpFunc func(args []any) (any, error)

type Evaluator struct {
	scope *Scope
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
	opSig, ok := v.(*OpSig)
	if !ok {
		if op.Inv != nil {
			return nil, fmt.Errorf("bad invocation on nonfunction variable")
		}
		v, err := e.Eval(v)
		if err != nil {
			return nil, fmt.Errorf("%v at var arg", err)
		}
		return v, nil
	}
	if op.Inv == nil {
		return op, nil
	}
	if n := len(op.Inv.Args); n > opSig.NArgs {
		return nil, fmt.Errorf("too many args for op %s (expected %d, found %d)", op.Id, opSig.NArgs, n)
	}
	args := make([]any, len(op.Inv.Args))
	for i, a := range op.Inv.Args {
		a, err := e.Eval(a)
		if err != nil {
			return nil, fmt.Errorf("%v at op arg", err)
		}
		args[i] = a
	}
	v, err := opSig.Func(args)
	if err != nil {
		return nil, err
	}
	if _, ok := v.(Op); !ok {
		return v, nil
	}
	op.Id = v.(Op).Id
	return e.evalOp(op)
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
