package foxmarks

import (
	"sort"
)

func isRenderWhitespace(r rune) bool {
	if r == ' ' || r == '\n' || r == '\t' {
		return true
	}

	return false
}


func cleanWhitespace(s string) string {
	i := 0
	j := len(s) - 1
	for isRenderWhitespace(rune(s[i])) {
		i++
	}

	for isRenderWhitespace(rune(s[j])) {
		j--
	}

	return s[i:j+1]
}

type BlockObjectRender struct {
	Object *BlockObject
	Offset int
	sOffset []int
	Render string
}

func (d *Document) Render() string {
	var str string

	for _, o := range d.Content {
		b := &BlockObjectRender{Object:o,Render:o.Content}
		str = str + b.renderBlock()
	}

	return str
}

func (b *BlockObjectRender) renderBlock() string {
	var str string
	set := BlockHTMLTags[b.Object.Type]

	if !set.CleanContent {
		if len(b.Object.Inlines) == 0 {
			str = set.Opener + cleanWhitespace(b.Object.Content)
		} else {
			b.renderInlines()
			
			str = set.Opener + cleanWhitespace(b.Render)
		}
	} else {
		str = set.Opener + ""
	}

	if len(b.Object.Blocks) != 0 {
		for _, o := range b.Object.Blocks {
			c := &BlockObjectRender{Object:o,Render:o.Content}
			str = str + c.renderBlock()
		}
	}

	return str + set.Closer
}

type InlineTag struct {
	Type InlineObjectType
	Content []string
	Tag InlineTagType
	Pos int
}

type InlineTagType int

const (
	Opener InlineTagType = 0
	Closer InlineTagType = 1
)

type InlineTags []*InlineTag

func (t InlineTags) Len() int	{ return len(t) }
func (t InlineTags) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

type InlineSort struct{ InlineTags }

func (s InlineSort) Less(i, j int) bool { return s.InlineTags[i].Pos < s.InlineTags[j].Pos }

func (b *BlockObjectRender) renderInlines() {
	var i []*InlineTag

	for _, t := range b.Object.Inlines {
		i = append(i, &InlineTag{
			Type: t.Type,
			Content: t.Content,
			Tag: Opener,
			Pos: t.StartPos,
		})
		i = append(i, &InlineTag{
			Type: t.Type,
			Content: t.Content,
			Tag: Closer,
			Pos: t.EndPos,
		})
	}

	sort.Sort(InlineSort{i})
	for _, n := range i {
		debug("Sorted tags:")
		debug(n.Type)
		debug(n.Pos)
	}

	/*
	for _, n := range i {
		var o int
		t := InlineHTMLTags[n.Type]

		switch n.Tag {
		case Opener:
			o = len(t.Opener) - t.OpLen
		case Closer:
			o = len(t.Closer) - t.OpLen
		}

		
		if p + 1 < len(i) {
			for _, n := range i[p+1:] {
				n.Pos = n.Pos + o
			}
		}
	}
	*/

	var o int
	for _, n := range i {
		var s string
		t := InlineHTMLTags[n.Type]
		
		switch n.Tag {
		case Opener:
			s = t.Opener(n)
		case Closer:
			s = t.Closer(n)
		}

		n.Pos = n.Pos + o
		o = o + len(s) - t.OpLen(n)
		b.Render = b.Render[0:n.Pos] + s + b.Render[n.Pos+t.OpLen(n):]

		debug(b.Render)
	}
}

type InlineHTMLTag func(*InlineTag) string
type InlineOpLen func(*InlineTag) int

type InlineHTMLTagset struct {
	Opener InlineHTMLTag
	Closer InlineHTMLTag
	OpLen InlineOpLen
}

var InlineHTMLTags map[InlineObjectType]InlineHTMLTagset = map[InlineObjectType]InlineHTMLTagset{
	StrongText: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "<strong>" },
		Closer: func(i *InlineTag) string { return "</strong>" },
		OpLen: func(i *InlineTag) int { return 2 },
	},
	EmText: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "<em>" },
		Closer: func(i *InlineTag) string { return "</em>" },
		OpLen: func(i *InlineTag) int { return 1 },
	},
	CodeSpan: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "<code>" },
		Closer: func(i *InlineTag) string { return "</code>" },
		OpLen: func(i *InlineTag) int { return 1 },
	},
	LinkRef: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "<a href=\"" + i.Content[1] + "\">" },
		Closer: func(i *InlineTag) string { return "</a>" },
		OpLen: func (i *InlineTag) int { return 1 },
	},
	Link: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "" },
		Closer: func(i *InlineTag) string { return "" },
		OpLen: func (i *InlineTag) int {
			if i.Tag == Opener {
				return len(i.Content[0]) + 2
			}

			return 0
		},
	},
	ImageRef: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "<img src=\"" + i.Content[1] + "\">" },
		Closer: func(i *InlineTag) string { return "" },
		OpLen: func (i *InlineTag) int {
			if i.Tag == Opener {
				return 2
			}	
			return 1 + len(i.Content[0])
		},
	},
	Image: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "" },
		Closer: func(i *InlineTag) string { return "" },
		OpLen: func (i *InlineTag) int {
			if i.Tag == Opener {
				return len(i.Content[0]) + 2
			}

			return 0
		},
	},
	LineBreak: InlineHTMLTagset{
		Opener: func(i *InlineTag) string { return "<br>" },
		Closer: func(i *InlineTag) string { return "" },
		OpLen: func(i *InlineTag) int {
			if len(i.Content) > 0 && i.Tag == Opener {
				return 1
			}
			return 0
		},
	},
}

type BlockHTMLRender struct {
	Opener string
	Closer string
	CleanContent bool
}

var BlockHTMLTags map[BlockObjectType]BlockHTMLRender = map[BlockObjectType]BlockHTMLRender{
	Paragraph: BlockHTMLRender{
		Opener: "<p>",
		Closer: "</p>",
	},
	List: BlockHTMLRender{
		Opener: "<ul>",
		Closer: "</ul>",
		CleanContent: true,
	},
	ListItem: BlockHTMLRender{
		Opener: "<li>",
		Closer: "</li>",
	},
	ThematicBreak: BlockHTMLRender{
		Opener: "",
		Closer: "<hr />",
		CleanContent: true,
	},
	Header1: BlockHTMLRender{
		Opener: "<h1>",
		Closer: "</h1>",
	},
	Header2: BlockHTMLRender{
		Opener: "<h2>",
		Closer: "</h2>",
	},
	Header3: BlockHTMLRender{
		Opener: "<h3>",
		Closer: "</h3>",
	},
	Header4: BlockHTMLRender{
		Opener: "<h4>",
		Closer: "</h4>",
	},
	Header5: BlockHTMLRender{
		Opener: "<h5>",
		Closer: "</h5>",
	},
	Header6: BlockHTMLRender{
		Opener: "<h6>",
		Closer: "</h6>",
	},
	CodeBlock: BlockHTMLRender{
		Opener: "<pre><code>",
		Closer: "</code></pre>",
	},
}
