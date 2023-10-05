package jsondsl

import "math"

// builtinOps lists builtin pure operations.
var builtinOps = map[string]*OpSig{
	"add":  {2, false, 1, add},
	"div":  {2, false, 1, div},
	"mean": {1, false, 1, mean},
	"mul":  {2, false, 1, mul},
	"sub":  {2, false, 1, sub},
	"sum":  {0, true, 1, sum},
}

func add(args []any) (any, error) {
	x, y := args[0], args[1]
	return x.(float64) + y.(float64), nil
}

func div(args []any) (any, error) {
	x, y := args[0], args[1]
	return x.(float64) / y.(float64), nil
}

func mean(args []any) (any, error) {
	vs := args[0].([]any)
	if len(vs) == 0 {
		return math.NaN(), nil
	}
	var sum float64
	for _, v := range vs {
		sum += v.(float64)
	}
	return sum / float64(len(vs)), nil
}

func mul(args []any) (any, error) {
	x, y := args[0], args[1]
	return x.(float64) * y.(float64), nil
}

func sub(args []any) (any, error) {
	x, y := args[0], args[1]
	return x.(float64) - y.(float64), nil
}

func sum(args []any) (any, error) {
	vs := args[0].([]any)
	var sum float64
	for _, v := range vs {
		sum += v.(float64)
	}
	return sum, nil
}
