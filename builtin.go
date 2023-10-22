package jsondsl

import (
	"fmt"
	"strconv"
)

// builtinOps lists builtin pure operations.
var builtinOps = map[string]OpFunc{
	"bind":   bind,
	"lambda": lambda,
}

func bind(scope *Scope, args []any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("bind expects 2 arguments: got %d", len(args))
	}
	var name string
	switch k := args[0].(type) {
	case *Op:
		if len(k.Args) != 0 {
			return nil, fmt.Errorf("not a valid name in arg 0 of bind: arg must be id or string")
		}
		name = k.Id
	case *String:
		// TODO(wes): Test contents for id conformity.
		var err error
		name, err = strconv.Unquote(k.QuotedContent)
		if err != nil {
			return nil, fmt.Errorf("failed to unquote arg 0 of bind: %v", err)
		}
	default:
		return nil, fmt.Errorf("not a valid name in arg 0 of bind: invalid type %T", k)
	}
	v, err := Eval(scope, args[1])
	if err != nil {
		return nil, err
	}
	scope.Bind(name, v)
	return nil, nil
}

func lambda(scope *Scope, args []any) (any, error) {
	switch len(args) {
	case 0:
		return func(*Scope, []any) (any, error) { return nil, nil }, nil
	case 1:
		return func(*Scope, []any) (any, error) { return args[0], nil }, nil
	default:
		return func(scope *Scope, as []any) (any, error) {
			if len(as) != len(args)-1 {
				return nil, fmt.Errorf("lambda expects %d argument, found %d", len(args)-1, len(as))
			}
			scope = scope.LocalScope()
			for i, a := range args[:len(args)-1] {
				v, ok := a.(*Op)
				if !ok || len(v.Args) != 0 {
					return nil, fmt.Errorf("not a valid variable in argument %d of lambda", i+1)
				}
				scope.Bind(v.Id, as[i])
			}
			return Eval(scope, args[len(args)-1])
		}, nil
	}
}
