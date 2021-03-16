package foxmarks

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"os"
	"text/scanner"
)

var Debug bool

func debug(i interface{}) {
	if Debug { 
		log.Println(i)
	}
}

type stack struct {
	l *list.List
}

func newStack() *stack {
	s := new(stack)
	s.l = list.New()

	return s
}

func (s *stack) push(i interface{}) {
	s.l.PushFront(i)
}

func (s *stack) pop() interface{} {
	return s.l.Remove(s.l.Front())
}

func (s *stack) head() *list.Element {
	return s.l.Front()
}

func (s *stack) len() int {
	return s.l.Len()
}

func (s *stack) remove(e *list.Element) {
	debug("removing from a stack")
	s.l.Remove(e)
}

func (s *stack) debugPrintStack() {
	if Debug {
		log.Println("printing contents of stack")
		for e := s.head() ; e != nil ; e = e.Next() {
			log.Println(e.Value)
		}
	}
}

// documentConstructor represents a struct that contains
// all the things you need to construct a document according to the state machine specification.
//
// It holds a Document object, a text scanner, and two queues for what block/inline objects are
// currently being processed.


type charScanner struct {
	Scanner *scanner.Scanner
	Cur rune
	Pos int
}

func newCharScanner(i io.Reader) *charScanner {
	c := new(charScanner)
	c.Scanner = new(scanner.Scanner)
	c.Scanner = c.Scanner.Init(i)
	c.Scanner.Mode = scanner.ScanStrings
	c.Scanner.Whitespace = 0
	c.Cur = c.Scanner.Scan()
	c.Pos = 0

	return c
}

func (c *charScanner) Scan() rune {
	c.Cur = c.Scanner.Scan()
	c.Pos++

	return c.Cur
}

func isWhiteSpace(r rune) bool {
	if r == '\n' || r == '\t' {
		return true
	}

	return false
}

type Document struct {
	Content    []*BlockObject // A document is comprised of only blocks - so, this makes sense for accessing.
	References []*Reference   // A document, however, can also have references.
}

type Reference struct {
	Label string
	Link  string
	Title string
}

type documentConstructor struct {
	Document *Document
	Scanner  *charScanner
	objStack *stack
	inlStack *stack
	actStack *stack
	HoldAction bool
}

// some consts to refer to what queue to use so we don't have have 2x the functions for each queue function
type dcstack int

const (
	objStack dcstack = 0
	inlStack dcstack = 1
	actStack dcstack = 2
)

func NewDocumentConstructor(i io.Reader) *documentConstructor {
	d := new(documentConstructor)
	d.Document = new(Document)
	d.Scanner = newCharScanner(i)
	d.objStack = newStack()
	d.inlStack = newStack()
	d.actStack = newStack()

	return d
}

func (d *documentConstructor) push(i interface{}) error {
	switch t := i.(type) {
	case *BlockObject:
		debug("pushed into objStack")
		d.objStack.push(i)
	case *InlineObject:
		debug("pushed into inlStack")
		d.inlStack.push(i)
	case SpecialAction:
		debug("pushed into actStack")
		d.actStack.push(i)
	default:
		return fmt.Errorf("not a valid document object, type %T", t)
	}

	return nil
}

func (d *documentConstructor) pop(t dcstack) {
	switch t {
	case objStack:
		d.objStack.pop()
	case inlStack:
		d.inlStack.pop()
	case actStack:
		// actStack should not be popped from, individual actions should instead
		// remove themselves from the stack. actStack is a stack because it minimizes
		// the possibility of the O(n) worst case, in case we have multiple actions occuring
		fmt.Println("warning: actStack should not be popped, let individual actions handle this")
	}
}

func (d *documentConstructor) currentHead(t dcstack) *list.Element {
	switch t {
	case objStack:
		return d.objStack.head()
	case inlStack:
		return d.inlStack.head()
	case actStack:
		return d.actStack.head()
	}

	return nil
}

func (d *documentConstructor) objectCur() *BlockObject {
	if e := d.currentHead(objStack) ; e != nil {
		return e.Value.(*BlockObject)
	}

	return nil
}

func (d *documentConstructor) inlineCur() *InlineObject {
	if e := d.currentHead(inlStack) ; e != nil {
		return e.Value.(*InlineObject)
	}

	return nil
}

func (d *documentConstructor) queueCheck(i interface{}) *list.Element {
	switch t := i.(type) {
	case BlockObjectType:
		for e := d.objStack.head() ; e != nil ; e = e.Next() {
			if e.Value.(*BlockObject).Type == t {
				return e
			}
		}
	case InlineObjectType:
		for e := d.inlStack.head() ; e != nil ; e = e.Next() {
			if e.Value.(*InlineObject).Type == t {
				return e
			}
		}
	}

	return nil
}


func (d *documentConstructor) performActions() {
	for a := d.actStack.head() ; a != nil ; {
		b := a.Next()
		a.Value.(SpecialAction)(d, a)
		a = b
	}
}

func (d *documentConstructor) closeObject(e *list.Element) {
	debug("object closing")
	switch t := e.Value.(type) {
	case *InlineObject:
		debug("inline object: setting end position")
		t.EndPos = d.Scanner.Pos - 1 // this appears to be a pattern
		d.objectCur().Inlines = append(d.objectCur().Inlines, t)
		debug(t)
		d.inlStack.remove(e)
	case *BlockObject:
		debug("block object: clearing stacks")
		debug(t)
		d.inlStack = newStack() // clear out the stack, as it is no longer needed
		// d.actStack = newStack() // most inline-related actions end on newlines, anyways
		if f := e.Next() ; f == nil {
			debug("root is document (no leading blocks), appending")
			d.Document.Content = append(d.Document.Content, t)
		} else {
			debug("root is block (leading block), appending")
			// most of the time, the objStack will be completely empty save for the head,
			// but otherwise if it isn't, that implies we're in a container block
			// so, we append the new block to the block above it in order to adhere
			// to the container block definition
			f.Value.(*BlockObject).Blocks = append(f.Value.(*BlockObject).Blocks, t)
		}
		d.objStack.debugPrintStack()
		d.objStack.remove(e)
	}
}

func Open(p string) (*documentConstructor, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	return NewDocumentConstructor(f), nil
}

// Construct is the base case of the document builder.
// It will scan through the document and append whatever it finds to the current block object,
// until a special character is found.
func (d *documentConstructor) Parse() *Document {
	for ; d.Scanner.Cur != scanner.EOF ; d.Scanner.Scan() {
		debug(string(d.Scanner.Cur))
		if d.actStack.head() != nil {
			d.performActions()
		}

		if d.objectCur() == nil && !isWhiteSpace(d.Scanner.Cur) {
			if a := d.checkSpecial(d.Scanner.Cur); a != nil && !d.HoldAction {
				d.Scanner.Pos = 0
				debug("creating object due to non-whitespace found")
				debug("special found")
				d.push(a)
				d.performActions()
			}

			// if no block object was created from the special, then just make a paragraph
			if d.objectCur() == nil && !d.HoldAction {
				debug("current character did not create a new block object - creating paragraph")
				d.Scanner.Pos = 0
				d.push(NewParagraph()) // the base block of markdown
			}
		}

		// you can't check the Initialized field,
		// if the object is nil
		if d.objectCur() != nil {
			if d.objectCur().Initialized {
				d.objectCur().appendContent(d.Scanner.Cur)

				if a := d.checkSpecial(d.Scanner.Cur); a != nil && !d.HoldAction {
					debug("special found")
					d.push(a)
				}
			}
		}

		if d.HoldAction {
			d.HoldAction = false
		}
	}

	debug("EOF - closing all remaining objects")

	if d.objStack.len() > 0 {
		d.closeAllObjects(d.objStack.head())
	}


	return d.Document
}
