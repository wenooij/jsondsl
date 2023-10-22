package jsondsl

import "fmt"

type Scope struct {
	Parent *Scope
	Vars   map[string]any
}

func BuiltinScope() *Scope {
	s := &Scope{Vars: make(map[string]any, len(builtinOps))}
	for id, op := range builtinOps {
		s.Bind(id, op)
	}
	return s
}

func (s *Scope) Reset(parent *Scope) {
	s.Parent = parent
	s.Vars = make(map[string]any)
}

func (s *Scope) Lookup(id string) (any, error) {
	for s != nil {
		if v, ok := s.Vars[id]; ok {
			return v, nil
		}
		s = s.Parent
	}
	return nil, fmt.Errorf("name %q not found", id)
}

func (s *Scope) Bind(id string, val any) (oldVal any, overwrote bool) {
	oldVal, overwrote = s.Vars[id]
	s.Vars[id] = val
	return oldVal, overwrote
}

func (s *Scope) LocalScope() *Scope {
	return &Scope{Parent: s, Vars: make(map[string]any)}
}
