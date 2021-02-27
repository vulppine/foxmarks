package foxmarks

// InlineObject represents an inline object within a block object.
// It consists of several things:
// - A start position, which indicates where in the content it should start
// - An end position, which indicates where in the content it should end.
//
// Inline objects can be several types of things (bold, URL, emphasis, etc),
// so it is a method interface.
type InlineObject struct {
	Type InlineObjectType
	StartPos int            // indicates the current starting position of the object relative to content
	EndPos int              // indicates the ending position of the object relative to content
	Content []string
	Closer Specials       // returns whatever closes the inline object
}

type InlineObjectType int

const (
	GenericInline InlineObjectType = iota
	StrongText
	LineBreak
	EmText
	LinkRef
	Link
	ImageRef
	Image
	CodeSpan
)

type InlineContent struct {
	HasContent bool
	Content    string
}

type Specials int

const (
	DoubleNewLine Specials = iota
	DoubleAsterisk
	NewLine
	Asterisk
	None
)

func (inl *InlineObject) appendContent(i interface{}) {
	switch t := i.(type) {
	case rune:
		inl.Content = append(inl.Content, string(t))
	case string:
		inl.Content = append(inl.Content, t)
	}
}

func NewInline(p int, t InlineObjectType) *InlineObject {
	return &InlineObject{
		Type: t,
		StartPos: p,
	}
}

func NewGenericInline(p int) *InlineObject {
	return &InlineObject{
		Type: GenericInline,
		StartPos: p,
	}
}

func NewStrongText(p int) *InlineObject {
	return &InlineObject{
		Type: StrongText,
		StartPos: p,
		Closer: DoubleAsterisk,
	}
}

func NewEmText(p int) *InlineObject {
	return &InlineObject{
		Type: EmText,
		StartPos: p,
		Closer: Asterisk,
	}
}

/*
func NewLink(p int) *InlineObject {
	return &InlineObject{
		Type: LinkText,
		StartPos: p,
	}
}
*/
