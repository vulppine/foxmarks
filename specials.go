package foxmarks

import (
	"container/list"
)

type actionConstructor func(*documentConstructor) SpecialAction

type SpecialAction func(*documentConstructor, *list.Element)

var specialRunes map[rune]actionConstructor = map[rune]actionConstructor{
	'\\': func(d *documentConstructor) SpecialAction {
		return func(d *documentConstructor, a *list.Element) {
			if d.blockCheck() {
				if d.Scanner.Cur == '\n' {
					d.objectCur().appendObject(&InlineObject{
						Type:     LineBreak,
						Content:  []string{"\\"},
						StartPos: d.Scanner.Pos - 1,
						EndPos:   d.Scanner.Pos,
					})
				}
			}

			d.HoldAction = true
			d.actStack.remove(a)
		}
	},
	' ': func(d *documentConstructor) SpecialAction {
		if d.blockCheck() {
			if l := d.queueCheck(List) ; l == nil {
				e := NewGenericInline(d.Scanner.Pos)

				return func(d *documentConstructor, a *list.Element) {
					if d.Scanner.Cur == ' ' {
						d.HoldAction = true
						d.push(func(d *documentConstructor) SpecialAction {
							return func(d *documentConstructor, a *list.Element) {
								if d.Scanner.Cur == ' ' {
									e.appendContent(d.Scanner.Cur)
									d.HoldAction = true
								} else if d.Scanner.Cur == '\n' {
									d.HoldAction = true
									d.push(func(d *documentConstructor) SpecialAction {
										return func(d *documentConstructor, a *list.Element) {
											if d.Scanner.Cur == ' ' {
												d.HoldAction = true
											} else {
												debug("closing hard break")
												e.EndPos = d.Scanner.Pos
												e.Type = LineBreak
												d.objectCur().appendObject(e)
												d.actStack.remove(a)
											}
										}
									}(d))
									d.actStack.remove(a)
								} else {
									d.objectCur().appendContent(e.Content)
									d.actStack.remove(a)
								}
							}
						}(d))
						d.actStack.remove(a)
					} else {
						// it really isn't anything
						d.actStack.remove(a)
					}
				}
			}
		} else {
			e := NewGeneric()
			p := 0

			d.push(e)
			return func(d *documentConstructor, a *list.Element) {
				if d.Scanner.Cur == ' ' && p < 4 {
					p++
					d.HoldAction = true
				} else if p >= 4 {
					e.Type = CodeBlock
					e.Initialized = true
					d.push(func(d *documentConstructor) SpecialAction {
						d.HoldAction = true
						var sscan bool
						var s int
						return func(d *documentConstructor, a *list.Element) {
							d.HoldAction = true
							if sscan && s < 4 {
								if d.Scanner.Cur == '\n' {
									e.appendContent(d.Scanner.Cur)
									return
								}

								if d.Scanner.Cur == ' ' {
									s++
								} else {
									// e := NewParagraph()
									d.HoldAction = false
									d.closeObject(d.currentHead(objStack))

									// d.push(e)
									d.actStack.remove(a)
								}
							} else if d.Scanner.Cur == '\n' {
								e.appendContent(d.Scanner.Cur)
								e.Initialized = false
								s = 0
								sscan = true
							} else if s == 4 {
								sscan = false
								e.Initialized = true
							}
						}
					}(d))
					d.actStack.remove(a)
				} else if d.Scanner.Cur != ' ' {
					switch d.Scanner.Cur {
					case '*', '-', '+':
						e.appendContent(d.Scanner.Cur)
						r := d.Scanner.Cur
						d.push(func(d *documentConstructor) SpecialAction {
							s := 0
							return func(d *documentConstructor, a *list.Element) {
								if d.Scanner.Cur == ' ' && s < 4 {
									e.appendContent(d.Scanner.Cur)
									s++
									d.HoldAction = true
								} else if d.Scanner.Cur != ' ' && s >= 1 {
									debug("new list - current prefix spaces: ")
									debug(p)
									d.pop(objStack)
									l := NewList(p, string(r), Unordered)
									d.push(d.newListController(nil, &p, l, string(d.Scanner.Cur)))
									d.actStack.remove(a)
								} else {
									e.Type = Paragraph
									e.Initialized = true
									e.appendContent(d.Scanner.Cur)
									d.actStack.remove(a)
								}
							}
						}(d))
						d.actStack.remove(a)
					default:
						e.Type = Paragraph
						e.Initialized = true
						e.appendContent(d.Scanner.Cur)
						d.actStack.remove(a)
					}
				}
			}
		}

		return nil
	},
	'`': func(d *documentConstructor) SpecialAction {
		if d.blockCheck() && d.objectCur().Initialized == true {
			d.push(NewInline(d.Scanner.Pos, CodeSpan))
			return func(d *documentConstructor, a *list.Element) {
				d.HoldAction = true
				if d.Scanner.Cur == '`' {
					if e := d.queueCheck(CodeSpan) ; e != nil {
						d.closeObject(e)
						e.Value.(*InlineObject).EndPos = d.Scanner.Pos 
					}

					d.actStack.remove(a)
				}
			}
		} else {
			e := NewGeneric()
			l := 0
			c := '`'
			d.push(e)
			return func(d *documentConstructor, a *list.Element) {
				if d.Scanner.Cur == c {
					l++
				} else if d.Scanner.Cur == '\n' && l >= 3 {
					v := 0 
					e.Type = CodeBlock
					e.Initialized = true
					e.Content = ""
					d.push(func() SpecialAction {
						return func(d *documentConstructor, a *list.Element) {
							d.HoldAction = true
							if d.Scanner.Cur == '\n' {
								e.Initialized = false
								d.push(func() SpecialAction {
									s := ""
									return func(d *documentConstructor, _a *list.Element) {
										d.HoldAction = true
										s = s + string(d.Scanner.Cur)
										if d.Scanner.Cur == '`' {
											v++
											debug(v)
										} else if d.Scanner.Cur == '\n' {
											debug(v)
											debug(l)
											if v >= l {
												debug("closing code fence")
												if e := d.queueCheck(CodeBlock) ; e != nil {
													d.closeObject(e)
												}

												d.actStack.remove(a)
												d.actStack.remove(_a)
											} else {
												e.Initialized = true
												e.Content = e.Content + s
												d.actStack.remove(_a)
											}
										} else {
											e.Initialized = true
											e.Content = e.Content + s
											d.actStack.remove(_a)
										}
									}
								}())
							}
						}
					}())
					d.actStack.remove(a)
				} else {
					e.Type = Paragraph
					e.Initialized = true
					d.actStack.remove(a)
				}
			}
		}

		return nil
	},
	'*': func(d *documentConstructor) SpecialAction {
		if d.blockCheck() {
			return strongOrEm(d, '*')
		} else {
			return func(d *documentConstructor, a *list.Element) {
				debug("potential thematic break detected")
				d.HoldAction = true
				d.push(listOrLine(d, a, '*', nil))

				d.actStack.remove(a)
			}
		}

		return nil // uh oh!
	},
	'#': func(d *documentConstructor) SpecialAction {
		if d.objectCur() == nil {
			d.Scanner.Pos = 0
			debug("creating header")
			lv := 0
			e := NewHeader()
			d.push(e)

			return func(d *documentConstructor, a *list.Element) {
				if d.Scanner.Cur == '#' && lv < 6 {
					e.appendContent(d.Scanner.Cur)
					d.HoldAction = true
					lv++
				} else if d.Scanner.Cur == ' ' {
					debug(lv)
					e.Initialized = true
					e.Content = ""
					e.Type = headerLevels[lv]
					d.actStack.remove(a)
				} else {
					debug("header level too high, converting to paragraph")
					e.Initialized = true
					e.Type = Paragraph // it no longer is a header
					d.actStack.remove(a)
				}
			}
		}

		return nil
	},
	'!': func(d *documentConstructor) SpecialAction {
		return func(d *documentConstructor, a *list.Element) {
			d.HoldAction = true
			if d.Scanner.Cur == '[' {
				d.push(d.referenceConstructor(Image))
			}

			d.actStack.remove(a)
		}
	},
	'[': func(d *documentConstructor) SpecialAction {
		if d.blockCheck() {
			if d.inlineCur() != nil {
				if d.inlineCur().Type == Image && d.inlineCur().Type == ImageRef {
					return nil
				}
			}

			debug("creating link")
			return d.referenceConstructor(Link)
			/*
			e := NewLink(d.Scanner.Pos)
			d.push(e)
			return func(d *documentConstructor, a *list.Element) {
				if e.Type == LinkText {
					if d.Scanner.Cur == ']' {
						debug("pushing an action to stop link construction")
						l := d.queueCheck(LinkText)
						d.push(func(d *documentConstructor) SpecialAction { // FUCKY
							return func(d *documentConstructor, _a *list.Element) {
								debug("checking if we're still in link mode...")
								if d.Scanner.Cur != '(' {
									debug("we are not, removing link")
									d.actStack.remove(a)
									d.inlStack.remove(l)
								} else {
									debug("we are, converting to final type")
									e.Type = Link
								}

								d.actStack.remove(_a)
							}
						}(d))
					}
				} else if e.Type == Link {
					if d.Scanner.Cur == ')' {
						debug("attempting to close inline link")
						if l := d.queueCheck(Link); l != nil {
							debug("closing inline link")
							d.closeObject(l)
						}
					} else {
						debug("constructing link")
						e.appendContent(d.Scanner.Cur)
					}
				}
			}
			*/

		}

		return nil
	},
	'\n': func(d *documentConstructor) SpecialAction {
		if e := d.objStack.head(); e != nil {
			switch e.Value.(*BlockObject).Type {
			case Paragraph, List:
				return func(d *documentConstructor, a *list.Element) {
					switch d.Scanner.Cur {
					case '\n':
						d.closeObject(e)
					case '*', '-':
						e.Value.(*BlockObject).Initialized = false
						d.push(listOrLine(d, nil, d.Scanner.Cur, nil))
					}

					d.actStack.remove(a)
				}
			case Header1, Header2, Header3, Header4, Header5, Header6:
				return func(d *documentConstructor, a *list.Element) {
					d.closeObject(e)
					d.actStack.remove(a)
				}
			}
		}

		return nil
	},
}

// MORE SPECIAL ACTIONS //

func (d *documentConstructor) referenceConstructor(t InlineObjectType) SpecialAction {
	var e *InlineObject
	var f *InlineObject
	var tr InlineObjectType
	var tf InlineObjectType

	switch t {
	case Link:
		e = NewInline(d.Scanner.Pos, LinkRef)
		f = NewInline(d.Scanner.Pos, Link)
		tr = LinkRef
		tf = Link
	case Image:
		// The checker for ! checks the exclamation mark first,
		// so we compensate by knocking the position back by one
		e = NewInline(d.Scanner.Pos - 1, ImageRef)
		f = NewInline(d.Scanner.Pos - 1, Image)
		tr = ImageRef
		tf = Image
	}

	d.push(e)
	d.push(f)
	var r string
	var k string
	c := true

	l := d.queueCheck(tr)
	m := d.queueCheck(tf)
	return func(d *documentConstructor, a *list.Element) {
		if c {
			if d.Scanner.Cur == ']' {
				debug("pushing an action to stop link construction")
				d.push(func(d *documentConstructor) SpecialAction { // FUCKY
					return func(d *documentConstructor, _a *list.Element) {
						debug("checking if we're still in link mode...")
						if d.Scanner.Cur != '(' {
							debug("we are not, removing link")
							d.actStack.remove(a)
							d.inlStack.remove(l)
							d.inlStack.remove(m)
						} else {
							debug("we are, checking ref now")
							d.closeObject(l)
							f.StartPos = d.Scanner.Pos
							c = false
						}

						e.appendContent(r)
						d.actStack.remove(_a)
					}
				}(d))
			} else {
				r = r + string(d.Scanner.Cur)
			}
		} else {
			if d.Scanner.Cur == ')' {
				debug("attempting to close reference")
				if m != nil {
					debug("closing inline link")
					e.appendContent(k)
					f.appendContent(k)
					d.closeObject(m)
					d.actStack.remove(a)
				}
			} else {
				if d.Scanner.Cur == '(' {
					return
				}

				debug("constructing link")
				k = k + string(d.Scanner.Cur)
				// e.appendContent(d.Scanner.Cur)
			}
		}
	}
}

func strongOrEm(d *documentConstructor, r rune) SpecialAction {
	return func(d *documentConstructor, a *list.Element) {
		if d.Scanner.Cur == r {
			if e := d.queueCheck(StrongText); e == nil {
				debug("making StrongText inline")
				d.HoldAction = true
				d.push(NewStrongText(d.Scanner.Pos - 1))
			} else {
				debug("closing StrongText inline")
				d.HoldAction = true
				d.closeObject(e)
			}
		} else {
			if e := d.queueCheck(EmText); e == nil {
				debug("making EmText inline")
				d.push(NewEmText(d.Scanner.Pos - 1))
			} else {
				debug("closing EmText inline")
				d.closeObject(e)
			}

		}

		d.actStack.remove(a)
	}
}

func listOrLine(d *documentConstructor, a *list.Element, t rune, e *BlockObject) SpecialAction {
	var l int

	if e == nil {
		e = NewGeneric()
		e.appendContent(d.Scanner.Cur)
		l = 1
	} else {
		l = 2
	}

	s := 0
	b := false
	return func(d *documentConstructor, a *list.Element) {
		d.HoldAction = true
		if d.Scanner.Cur == '\n' && l >= 3 {
			debug("creating and closing thematic break")
			e.Type = ThematicBreak
			e.Content = ""

			if d.objectCur() != nil {
				debug("closing all objects")
				d.closeAllObjects(d.currentHead(objStack))
			}

			d.push(e)
			d.closeObject(d.currentHead(objStack))
			d.actStack.remove(a)
			return
		}

		if d.Scanner.Cur == t {
			debug(l)
			e.appendContent(d.Scanner.Cur)
			l++
			b = true
		} else if d.Scanner.Cur != ' ' {
			b = true
			if s >= 1 {
				debug("appending as a new list")
				l := NewList(0, string(t), Unordered)
				e.appendContent(d.Scanner.Cur)
				if d.objectCur() != nil {
					debug("closing all objects")
					d.closeAllObjects(d.currentHead(objStack))
				}
				d.push(d.newListController(nil, nil, l, e.Content[1:]))
				d.actStack.remove(a)
			} else {
				if p := d.queueCheck(Paragraph) ; p != nil {
					debug("appending to paragraph")
					p.Value.(*BlockObject).Content = p.Value.(*BlockObject).Content + string('\n') + e.Content
					d.actStack.remove(a)
				} else {
					debug("appending as paragraph")
					e.Type = Paragraph
					e.Initialized = true
					d.push(e)
					d.actStack.remove(a)
				}
			}
		} else if !b {
			debug(s)
			e.appendContent(d.Scanner.Cur)
			s++
		}
	}
}

func (d *documentConstructor) newListController(phold *bool, pre *int, l *BlockObject, c ...string) SpecialAction {
	if pre == nil {
		pre = new(int)
		*pre = 0
	}

	var validOp bool
	var hold bool
	// var tight bool
	var nl int
	var li *BlockObject
	var act SpecialAction

	d.push(l)
	le := d.queueCheck(List)
	lattrib := l.GetListAttrib()

	closeList := func(a *list.Element) {
		for el := d.currentHead(objStack) ; el != le ; {
			ele := el.Next()
			d.closeObject(el)
			el = ele
		}

		d.closeObject(le)
		d.objStack.debugPrintStack()
		d.actStack.remove(a)
	}

	NewItem := func() (*BlockObject, SpecialAction) {
		e := NewListItem()
		d.Scanner.Pos = -1
		a := func(d *documentConstructor, a *list.Element) {
			e.Initialized = true
			if d.Scanner.Cur == '\n' {
				*pre = 0
				hold = false
				e.Initialized = false
				d.actStack.remove(a)
			}
		}

		return e, a
	}

	/*
	tightToLoose := func(l *BlockObject) {
		for _, li := range l.Blocks {
			p := NewParagraph()
			b := []*BlockObject{p}
			b = append(b, li.Blocks...)
			p.Content = li.Content

			li.Content = ""
			li.Blocks = b
		}
	}
	*/

	validListOp := func(o rune, a *list.Element) SpecialAction {
		if lattrib.Order == Unordered {
			if string(o) == lattrib.Op {
				return func(d *documentConstructor, _a *list.Element) {
					d.HoldAction = true
					if d.Scanner.Cur == ' ' {
						validOp = true
						hold = false
					} else {
						if phold != nil {
							*phold = false
							debug("closing list and passing up")
						} else {
							d.HoldAction = false
							debug("at root of list - closing and handing back to Constructor")
						}

						closeList(a)
					}

					d.actStack.remove(_a)
				}
			} else {
				// hand it off to the previous list,
				// or otherwise just close the list
				// and let the Constructor deal with it
				if phold != nil {
					*phold = false
					debug("closing list and passing up")
				} else {
					d.HoldAction = false
					debug("at root of list - closing and handing back to Constructor")
				}

				closeList(a)
			}
		} else {
			debug("ordered lists are not implemented yet")
		}

		return nil
	}

	listItemOrLine := func(a *list.Element, e *BlockObject, act SpecialAction) SpecialAction {
		var s int
		var b bool
		l := 1

		return func(d *documentConstructor, _a *list.Element) {
			d.HoldAction = true
			if d.Scanner.Cur == '\n' && l >= 3 {
				debug("creating thematic break, closing list")
				e.Type = ThematicBreak
				e.Content = ""

				closeList(a)
				d.actStack.remove(a)
				d.actStack.remove(_a)

				d.push(e)
				d.closeObject(d.currentHead(objStack))
				return
			}

			if string(d.Scanner.Cur) == lattrib.Op {
				e.appendContent(d.Scanner.Cur)
				l++
				b = true
			} else if d.Scanner.Cur != ' ' {
				d.HoldAction = false
				b = true
				if s >= 1 {
					debug("appending as list item")
					e.Content = e.Content[1:] + string(d.Scanner.Cur)
					d.push(e)
					d.push(act)
					d.actStack.remove(_a)
				} else {
					debug("creating paragraph, closing list")
					e.Type = Paragraph
					e.Initialized = true

					closeList(a)
					d.actStack.remove(a)
					d.actStack.remove(_a)

					d.push(e)
				}
			} else if !b {
				e.appendContent(d.Scanner.Cur)
				s++
			}
		}
	}

	if len(c) > 0 {
		e, a := NewItem()
		li = e

		for _, s := range c {
			e.appendContent(s)
		}

		hold = true
		d.push(e)
		d.push(a)
	}

	return func(d *documentConstructor, a *list.Element) {
		if phold != nil {
			// debug("currently holding a parent list")
		}

		if hold {
			return
		} else {
			debug("the current list is: ")
			debug(l)
			debug("refreshing list element")
			le = d.queueCheck(List)
			debug(le.Value)
			if le.Value.(*BlockObject) != l {
				panic("could not get list element, aborting")
			}
			d.HoldAction = true
		}

		if validOp {
			debug("list operator was valid, pushing")
			validOp = false
			hold = true

			li, act = NewItem()
			d.push(li)
			d.push(act)
		} else if *pre >= lattrib.Pre + len(lattrib.Op) && d.Scanner.Cur != ' ' {
			hold = true

			switch {
			case isListChar(d.Scanner.Cur):
				d.push(func() SpecialAction {
					c := string(d.Scanner.Cur)
					return func(d *documentConstructor, a *list.Element) {
						d.HoldAction = true
						if d.Scanner.Cur == ' ' {
							l := NewList(*pre, c, Unordered)
							*pre = 0

							debug("creating new list")
							d.push(d.newListController(&hold, pre, l, ""))
						} else {
							d.HoldAction = false
							debug("appending as content to current list item")
							c = c + string(d.Scanner.Cur)
							li.Content = li.Content + c
							li.Initialized = true
							d.push(act)
						}

						d.actStack.remove(a)
					}
				}())
			default:
				debug("appending as content to current list item")
				d.HoldAction = false
				li.appendContent('\n')
				li.Initialized = true
				d.push(act)
			}
		} else if *pre >= lattrib.Pre && d.Scanner.Cur != ' ' {
			if lattrib.Pre == 0 {
				if d.Scanner.Cur == '\n' {
					nl++
				} else {
					debug("checking if character is thematic break or list item - otherwise terminating")
					if d.objectCur().Type == ListItem {
					debug("closing list item")
						d.closeObject(d.currentHead(objStack))
					}

					e, act := NewItem()
					e.appendContent(d.Scanner.Cur)

					hold = true
					d.push(listItemOrLine(a, e, act))
				}
			} else {
				debug("checking if list operator is valid for new list item")
				if d.objectCur().Type == ListItem {
					debug("closing list item")
					d.closeObject(d.currentHead(objStack))
				}

				hold = true
				d.push(validListOp(d.Scanner.Cur, a))
			}
		} else if d.Scanner.Cur == ' ' {
			*pre++
			debug("current prefix spaces: ")
			debug(*pre)
		} else if d.Scanner.Cur == '\n' {
			nl++
		} else {
			if phold != nil {
				*phold = false
				debug("closing list and passing up - prefix spaces: ")
				debug(*pre)
			} else {
				d.HoldAction = false
				debug("at root of list - closing and handing back to Constructor")
			}

			closeList(a)
			d.actStack.remove(a)
		}
	}


	return nil
}


// recursively closes all objects in a stack,
// in case it is needed
func (d *documentConstructor) closeAllObjects(l *list.Element) {
	for l != nil {
		m := l.Next()
		d.closeObject(l)
		l = m
	}
}

// CHECKERS //

func isLetter(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}

	if r >= 'A' && r <= 'Z' {
		return true
	}

	return false
}

func isNumber(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}

	return false
}

func isListChar(r rune) bool {
	switch r {
	case '*', '-', '+':
		return true
	}

	return false
}

// checkSpecial only checks for if a unicode rune is a special syntax rune
// this should NOT be constructing types or anything like that,
// that should be deferred to a different function entirely
// I feel like (*CharScanner).checkSpecial(*DocumentCostructor) would be
// a little more readable, but this saves a bit of pointer usage.
func (d *documentConstructor) checkSpecial(r rune) SpecialAction {
	a, t := specialRunes[r] // is it a special rune?
	if t {
		return a(d)
	}

	return nil // return the nil value for a special, None
}

func (d *documentConstructor) blockCheck() bool {
	if d.objectCur() == nil && d.objStack.len() == 0 {
		return false
	}

	return true
}

// creates a paragraph if the current special being acted on
// is an inline object, and there is no pre-existing object either
// in the current selection, or in the stack
func (d *documentConstructor) inlineBlockCheck() {
	if !d.blockCheck() {
		d.push(NewParagraph())
	}
}
