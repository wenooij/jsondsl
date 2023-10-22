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
	Null struct{ NullPos Pos }
	Bool struct {
		LitPos  Pos
		Literal bool
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
		Elements []ListElem[Value]
		RBrack   Pos
	}
	Object struct {
		LBrace  Pos
		Members []ListElem[*Member]
		RBrace  Pos
	}
	Member struct {
		Key   Value // Key excluding *Array and *Object.
		Colon Pos
		Value Value
	}
	Ident struct {
		NamePos Pos
		Name    string
	}
	Operator struct {
		Id   *Ident
		Args []*OperatorArgs
	}
	OperatorArgs struct {
		LParen    Pos
		ValueList []ListElem[Value]
		RParen    Pos
	}
	ListElem[E Node] struct {
		Value E
		Comma Pos // CommaPos denotes the trailing comma, if any.
	}
)

func (a *Null) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.NullPos
}
func (a *Bool) Pos() Pos {
	if a == nil {
		return NoPos
	}
	return a.LitPos
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
	return a.Id.Pos()
}

func (*Null) val()     {}
func (*Bool) val()     {}
func (*Number) val()   {}
func (*String) val()   {}
func (*Array) val()    {}
func (*Object) val()   {}
func (*Operator) val() {}
