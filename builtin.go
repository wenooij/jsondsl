package jsondsl

import "fmt"

// builtinOps lists builtin pure operations.
var builtinOps = map[string]OpFunc{
	"bind":   bind,
	"lambda": lambda,
}

func bind(scope *Scope, args []any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("bind expects 2 arguments: got %d", len(args))
	}
	k, ok := args[0].(Op)
	if !ok || k.Inv != nil {
		return nil, fmt.Errorf("not a valid variable in argument 0 of bind")
	}
	v, err := Eval(scope, args[1])
	if err != nil {
		return nil, err
	}
	scope.Bind(k.Id, v)
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
				v, ok := a.(Op)
				if !ok || v.Inv != nil {
					return nil, fmt.Errorf("not a valid variable in argument %d of lambda", i+1)
				}
				scope.Bind(v.Id, as[i])
			}
			return Eval(scope, args[len(args)-1])
		}, nil
	}
}
