package ui

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// editor is the text field where the user writes messages and commands.
type editor struct {
	// text contains the written runes. An empty slice means no text is written.
	text []rune

	// textWidth[i] contains the width of string(text[:i]). Therefore
	// len(textWidth) is always strictly greater than 0 and textWidth[0] is
	// always 0.
	textWidth []int

	// cursorIdx is the index in text of the placement of the cursor, or is
	// equal to len(text) if the cursor is at the end.
	cursorIdx int

	// offsetIdx is the number of elements of text that are skipped when
	// rendering.
	offsetIdx int

	// width is the width of the screen.
	width int
}

func newEditor(width int) editor {
	return editor{
		text:      []rune{},
		textWidth: []int{0},
		width:     width,
	}
}

func (e *editor) Resize(width int) {
	if width < e.width {
		e.cursorIdx = 0
		e.offsetIdx = 0
	}
	e.width = width
}

func (e *editor) IsCommand() bool {
	return len(e.text) != 0 && e.text[0] == '/'
}

func (e *editor) TextLen() int {
	return len(e.text)
}

func (e *editor) PutRune(r rune) {
	e.text = append(e.text, ' ')
	copy(e.text[e.cursorIdx+1:], e.text[e.cursorIdx:])
	e.text[e.cursorIdx] = r

	rw := runewidth.RuneWidth(r)
	tw := e.textWidth[len(e.textWidth)-1]
	e.textWidth = append(e.textWidth, tw+rw)
	for i := e.cursorIdx + 1; i < len(e.textWidth); i++ {
		e.textWidth[i] = rw + e.textWidth[i-1]
	}

	e.Right()
}

func (e *editor) RemRune() (ok bool) {
	ok = 0 < e.cursorIdx
	if !ok {
		return
	}

	// TODO avoid looping twice
	rw := e.textWidth[e.cursorIdx] - e.textWidth[e.cursorIdx-1]
	for i := e.cursorIdx; i < len(e.textWidth); i++ {
		e.textWidth[i] -= rw
	}
	copy(e.textWidth[e.cursorIdx:], e.textWidth[e.cursorIdx+1:])
	e.textWidth = e.textWidth[:len(e.textWidth)-1]

	copy(e.text[e.cursorIdx-1:], e.text[e.cursorIdx:])
	e.text = e.text[:len(e.text)-1]
	e.Left()
	return
}

func (e *editor) Flush() (content string) {
	content = string(e.text)
	e.text = e.text[:0]
	e.textWidth = e.textWidth[:1]
	e.cursorIdx = 0
	e.offsetIdx = 0
	return
}

func (e *editor) Right() {
	if e.cursorIdx == len(e.text) {
		return
	}
	e.cursorIdx++
	if e.width < e.textWidth[e.cursorIdx]-e.textWidth[e.offsetIdx] {
		e.offsetIdx += 16
		max := len(e.text) - 1
		if max < e.offsetIdx {
			e.offsetIdx = max
		}
	}
}

func (e *editor) Left() {
	if e.cursorIdx == 0 {
		return
	}
	e.cursorIdx--
	if e.cursorIdx <= e.offsetIdx {
		e.offsetIdx -= 16
		if e.offsetIdx < 0 {
			e.offsetIdx = 0
		}
	}
}

func (e *editor) Draw(screen tcell.Screen, y int) {
	st := tcell.StyleDefault

	x := 0
	i := e.offsetIdx

	for i < len(e.text) && x < e.width {
		r := e.text[i]
		screen.SetContent(x, y, r, nil, st)
		x += runewidth.RuneWidth(r)
		i++
	}

	for x < e.width {
		screen.SetContent(x, y, ' ', nil, st)
		x++
	}

	curStart := e.textWidth[e.cursorIdx] - e.textWidth[e.offsetIdx]
	curEnd := curStart + 1
	if e.cursorIdx+1 < len(e.textWidth) {
		curEnd = e.textWidth[e.cursorIdx+1] - e.textWidth[e.offsetIdx]
	}
	for x := curStart; x < curEnd; x++ {
		screen.ShowCursor(x, y)
	}
}
