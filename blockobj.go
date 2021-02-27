package foxmarks

import (
	"fmt"
	"strconv"
	"strings"
)

// BlockObject is an interface for all kinds of block objects.
// In Markdown, all formatting is based around strings -
// so strings are the primitive in Markdown.

type BlockObject struct {
	Type BlockObjectType
	Initialized bool           // indicates if a block object is currently being initalized or not
	Content string          // the content of a block object, in string format. This is the raw string parsed by a constructor.
	Inlines []*InlineObject // whatever inline objects a block object may have
	Blocks []*BlockObject // some blocks can contain more blocks - if so, then a blockobject will contain these blocks.
}

type BlockObjectType int

const (
	Generic BlockObjectType = iota // generics are placeholders
	Paragraph
	Header1
	Header2
	Header3
	Header4
	Header5
	Header6
	ThematicBreak
	List
	ListItem
	CodeBlock
)

func (b *BlockObject) appendContent(i interface{}) error {
	switch v := i.(type) {
	case rune:
		b.Content = b.Content + string(v)
	case string:
		b.Content = b.Content + v
	default:
		return fmt.Errorf("invalid type to append to BlockObject: %T", v)
	}

	return nil
}

func (b *BlockObject) appendObject(i interface{}) error {
	switch o := i.(type) {
	case *BlockObject:
		b.Blocks = append(b.Blocks, o)
	case *InlineObject:
		b.Inlines = append(b.Inlines, o)
	default:
		return fmt.Errorf("invalid object to append to BlockObject: %T", o)
	}

	return nil
}

func NewBlockObject(t BlockObjectType, i bool) *BlockObject {
	return &BlockObject{
		Type: t,
		Initialized: i,
	}
}

func NewGeneric() *BlockObject {
	return &BlockObject{
		Type: Generic,
		Initialized: false,
	}
}

func NewParagraph() *BlockObject {
	return &BlockObject{
		Type: Paragraph,
		Initialized: true,
	}
}

func NewHeader() *BlockObject {
	return NewGeneric()
}

func NewThematicBreak() *BlockObject {
	return &BlockObject{
		Type: ThematicBreak,
		Initialized: false,
	}
}

// listOffsets is a way to encode how many 'spaces' a list needs
// before it is considered a part of a list structure
//
// pre indicates how many prefix spaces there are
// op indicates the length of the operator
// suf indicates how many suffix spaces there are
//
// for example:
//
// [*  ] would have listOffsets pre: 0, op: 2, suf: 1
// [ -     ] would have listOffsets pre: 1, op: 2, suf: 3
// [  1.   ] would have liftOffsets pre: 2, op: 3, suf: 2
//
// if a ListItem immediately precedes a newline char,
// there can only be precursor spaces equivalent to:
// - for a new item: pre
// - to continue the previous item: pre + op + suf
// relevant to the root/first item in the list
//
// otherwise, it automatically closes the entire list
//
// list examples:
// ```
// * [pre 0, op 2]
// * [pre 0, op 2]
// * [pre 0, op 2]
// ```
//
// ```
// * [pre 0, op 2, suf 0]
//   * [pre 2, op 2, suf 0]
//     * [pre 4, op 2, suf 0]
// ```

type ListOrder int

const (
	Unordered ListOrder = 0
	Ordered ListOrder = 1
)

type ListAttribs struct {
	Pre int
	Op string
	Order ListOrder
}

// listOffsets are encoded in the List's content as fieldless CSV
func (b *BlockObject) GetListAttrib() ListAttribs {
	if b.Type != List {
		return ListAttribs{}
	}

	var a ListAttribs
	v := strings.Split(b.Content,",")

	a.Pre, _ = strconv.Atoi(v[0])
	a.Op = v[1]
	switch v[2] {
	case "u":
		a.Order = Unordered
	case "o":
		a.Order = Ordered
	}

	return a
}

func NewList(p int, o string, r ListOrder) *BlockObject {
	l := NewBlockObject(List, false)

	l.Content = strings.Join(
		[]string{
			strconv.Itoa(p),
			o,
			func () string {
				switch r {
				case Unordered:
					return "u"
				case Ordered:
					return "o"
				}

				return ""
			}(),
		},
		",",
	)

	return l
}

func NewListItem() *BlockObject {
	return NewBlockObject(ListItem, false)
}

var headerLevels map[int]BlockObjectType = map[int]BlockObjectType{
	1: Header1,
	2: Header2,
	3: Header3,
	4: Header4,
	5: Header5,
	6: Header6,
}
