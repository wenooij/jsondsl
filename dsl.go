package jsondsl

type Pos int

const NoPos Pos = -1

type Node interface {
	Pos() Pos
}

type Value interface {
	Node
	val()
}

type (
	Null struct{ ValuePos Pos }
	Bool struct {
		ValuePos Pos
		Value    bool
	}
	Number struct {
		LitPos  Pos
		Literal string
	}
	String struct {
		Quote         Pos
		QuotedContent string
	}
	Array struct {
		LBrack   Pos
		Elements []Value
		RBrack   Pos
	}
	Object struct {
		LBrace  Pos
		Members []*Member
		RBrace  Pos
	}
	Member struct {
		Key   *String
		Colon Pos
		Value Value
	}
	Ident struct {
		NamePos Pos
		Name    string
	}
	Operator struct {
		Op     *Ident
		LParen Pos
		Args   []Value
		RParen Pos
	}
)

func (a *Null) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.ValuePos
}
func (a *Bool) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.ValuePos
}
func (a *Number) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.LitPos
}
func (a *String) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.Quote
}
func (a *Array) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.LBrack
}
func (a *Member) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.Key.Pos()
}
func (a *Object) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.LBrace
}
func (a *Ident) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.NamePos
}
func (a *Operator) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.Op.Pos()
}

func (*Null) val()     {}
func (*Bool) val()     {}
func (*Number) val()   {}
func (*String) val()   {}
func (*Ident) val()    {}
func (*Array) val()    {}
func (*Object) val()   {}
func (*Operator) val() {}
