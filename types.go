package jsondsl

// TypeName returns the name of the jsondsl type or "unknown".
func TypeName(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case float64:
		return "number"
	case string:
		return "string"
	case *Op:
		return "op"
	case []any:
		return "array"
	case map[any]any:
		return "object"
	default:
		return "unknown"
	}
}

// AsBool interprets v as a boolean.
func AsBool(v any) bool {
	switch v := v.(type) {
	case nil:
		return false
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	case *Op:
		return v != nil
	case []any:
		return len(v) != 0
	case map[any]any:
		return len(v) != 0
	default:
		return false
	}
}

func IsNull(v any) bool {
	switch v.(type) {
	case nil:
		return true
	default:
		return false
	}
}
