package jsondsl

type Scope struct {
	Parent *Scope
	Vars   map[string]any
}

func GlobalScope() *Scope {
	return &Scope{Vars: make(map[string]any)}
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

func (s *Scope) LookupOp(id string) (OpFunc, bool) {
	if v, ok := s.Lookup(id); ok {
		if sig, ok := v.(OpFunc); ok {
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

func (s *Scope) LocalScope() *Scope {
	return &Scope{Parent: s, Vars: make(map[string]any)}
}
