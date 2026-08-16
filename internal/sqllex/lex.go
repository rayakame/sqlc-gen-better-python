package sqllex

import "strings"

// Kind classifies a scanned token.
type Kind uint8

const (
	// KindText is ordinary SQL: copied as-is, holds nothing bindable.
	KindText Kind = iota
	// KindSkipped is a string literal, quoted identifier or comment. It is
	// copied as-is too, but is called out because a placeholder inside one
	// does not bind.
	KindSkipped
	// KindPlaceholder is one bindable slot.
	KindPlaceholder
	// KindSliceMarker is a /*SLICE:name*/ marker and the placeholder it
	// binds, which sqlc always emits adjacent.
	KindSliceMarker
)

// sliceMarkerPrefix opens the marker sqlc leaves for a sqlc.slice parameter.
const sliceMarkerPrefix = "/*SLICE:"

// dashComment is the line-comment introducer both engines share.
const dashComment = "--"

// decimalBase is the radix of a numbered placeholder's index.
const decimalBase = 10

// Token is one scanned span of the input, [Start, End).
type Token struct {
	Kind  Kind
	Start int
	End   int
	// Name is the sqlc.slice name of a KindSliceMarker token.
	Name string
	// Number is the explicit index of a numbered placeholder (sqlite's ?N),
	// or 0 when the placeholder carries none.
	Number int
	// MarkerEnd splits a KindSliceMarker: [Start, MarkerEnd) is the comment,
	// [MarkerEnd, End) the placeholder it binds. Zero for other kinds.
	MarkerEnd int
}

// Slot is a bindable position in the text, in text order. Name and Marker are
// empty for a plain placeholder; for a sqlc.slice they carry the raw name and
// the marker's exact text, which is what generated code has to replace.
type Slot struct {
	Name   string
	Marker string
	// Number is the explicit index of a numbered placeholder (sqlite's ?N),
	// or 0 when the placeholder carries none. SQLite binds ?N to slot N and
	// a bare ? to one past the highest slot seen so far, so the two spell
	// different bindings for the same text position.
	Number int
}

// Scan splits sql into tokens under the given dialect. Unterminated strings,
// identifiers and comments consume the rest of the input, matching how sqlc's
// own parsers recover.
func Scan(sql string, d Dialect) []Token {
	var tokens []Token
	text := 0
	flush := func(end int) {
		if end > text {
			tokens = append(tokens, Token{Kind: KindText, Start: text, End: end, Name: "", Number: 0, MarkerEnd: 0})
		}
	}
	skip := func(start, end int) {
		flush(start)
		tokens = append(tokens, Token{Kind: KindSkipped, Start: start, End: end, Name: "", Number: 0, MarkerEnd: 0})
		text = end
	}

	for i := 0; i < len(sql); {
		rest := sql[i:]
		switch {
		case strings.HasPrefix(rest, sliceMarkerPrefix):
			end := scanBlockComment(sql, i)
			// Only an immediately following placeholder binds to the marker;
			// anything else leaves an ordinary comment behind. A dialect
			// without a placeholder can bind nothing at all.
			token := end + len(d.placeholder)
			if d.placeholder != "" && token <= len(sql) && sql[end:token] == d.placeholder {
				flush(i)
				tokens = append(tokens, Token{
					Kind:      KindSliceMarker,
					Start:     i,
					End:       token,
					Name:      d.sliceName(sql[i+len(sliceMarkerPrefix) : end]),
					Number:    0,
					MarkerEnd: end,
				})
				i, text = token, token

				continue
			}
			skip(i, end)
			i = end
		case d.liveVersionComments && strings.HasPrefix(rest, "/*!"):
			// The body is executable SQL, so only the opener is consumed and
			// the closing */ falls through as ordinary text.
			i += len("/*!")
		case strings.HasPrefix(rest, "/*"):
			end := scanBlockComment(sql, i)
			skip(i, end)
			i = end
		case strings.HasPrefix(rest, dashComment) && (!d.dashNeedsGap || isCommentGap(sql, i+len(dashComment))):
			end := scanLineEnd(sql, i)
			skip(i, end)
			i = end
		case d.hashComments && rest[0] == '#':
			end := scanLineEnd(sql, i)
			skip(i, end)
			i = end
		case rest[0] == '\'' || rest[0] == '"' ||
			(d.backtickIdents && rest[0] == '`'):
			// Backslash escapes never apply inside backticks.
			end := scanQuoted(sql, i, rest[0], d.backslashEscapes && rest[0] != '`')
			skip(i, end)
			i = end
		case d.bracketIdents && rest[0] == '[':
			end := scanBracket(sql, i)
			skip(i, end)
			i = end
		case d.escaped != "" && strings.HasPrefix(rest, d.escaped):
			i += len(d.escaped)
		// The guard is what stops a zero-value Dialect: an empty prefix
		// matches everywhere, and the scan would never advance.
		case d.placeholder != "" && strings.HasPrefix(rest, d.placeholder):
			flush(i)
			end := i + len(d.placeholder)
			number := 0
			if d.numbered {
				digits := end
				for end < len(sql) && sql[end] >= '0' && sql[end] <= '9' {
					end++
				}
				// Overflow cannot happen in practice: sqlc rejects gaps, so
				// the highest index it emits is the parameter count.
				for _, b := range []byte(sql[digits:end]) {
					number = number*decimalBase + int(b-'0')
				}
			}
			tokens = append(tokens, Token{Kind: KindPlaceholder, Start: i, End: end, Name: "", Number: number, MarkerEnd: 0})
			i, text = end, end
		default:
			i++
		}
	}
	flush(len(sql))

	return tokens
}

// Placeholder returns one bindable placeholder as it appears in this
// dialect's text. Emitters use it so the token they write and the token the
// scanner looks for can never be two different literals.
func (d Dialect) Placeholder() string {
	return d.placeholder
}

// SliceMarker returns the marker sqlc leaves for a sqlc.slice parameter,
// together with the placeholder it binds. Scanning reports the marker's
// actual text; this rebuilds it for the one caller that has a name but no
// scanned text to point at.
func (d Dialect) SliceMarker(name string) string {
	return sliceMarkerPrefix + name + "*/" + d.placeholder
}

// Slots reports the bindable positions of sql in text order.
func Slots(sql string, d Dialect) []Slot {
	var slots []Slot
	for _, token := range Scan(sql, d) {
		switch token.Kind {
		case KindPlaceholder:
			slots = append(slots, Slot{Name: "", Marker: "", Number: token.Number})
		case KindSliceMarker:
			slots = append(slots, Slot{Name: token.Name, Marker: sql[token.Start:token.End], Number: 0})
		case KindText, KindSkipped:
		}
	}

	return slots
}

// sliceName recovers the sqlc.slice name from a marker body. In text whose
// literals were doubled by the rewriter the name is doubled too, so the
// dialect that describes such text undoes it - the name has to match the
// parameter sqlc reported, not the escaped spelling.
func (d Dialect) sliceName(body string) string {
	name := strings.TrimSuffix(body, "*/")
	if d.escaped == "" {
		return name
	}

	return strings.ReplaceAll(name, d.escaped, d.escaped[:len(d.escaped)/2])
}

// isCommentGap reports whether the byte at i lets a preceding "--" start a
// comment: MySQL wants whitespace, a control character, or end of input.
func isCommentGap(sql string, i int) bool {
	return i >= len(sql) || sql[i] <= ' '
}

// scanLineEnd returns the index of the \n ending a line comment, or the end
// of the input. The terminator is left for the caller: it is ordinary text.
// Both MySQL and SQLite end line comments only at \n, so a bare \r stays
// comment text.
func scanLineEnd(sql string, i int) int {
	end := strings.IndexByte(sql[i:], '\n')
	if end == -1 {
		return len(sql)
	}

	return i + end
}

// scanQuoted returns the index after a quoted region starting at i. A doubled
// quote is always an escape; a backslash escapes the next byte only where the
// dialect says so.
func scanQuoted(sql string, i int, quote byte, escapes bool) int {
	j := i + 1
	for j < len(sql) {
		switch {
		case escapes && sql[j] == '\\' && j+1 < len(sql):
			j += 2
		case sql[j] != quote:
			j++
		case j+1 < len(sql) && sql[j+1] == quote:
			j += 2
		default:
			return j + 1
		}
	}

	return len(sql)
}

// scanBracket returns the index after a [quoted identifier]. SQLite has no
// escape inside brackets: the first ] ends it.
func scanBracket(sql string, i int) int {
	end := strings.IndexByte(sql[i:], ']')
	if end == -1 {
		return len(sql)
	}

	return i + end + 1
}

// scanBlockComment returns the index after a /* */ comment starting at i.
// Neither engine nests them: the first */ wins.
func scanBlockComment(sql string, i int) int {
	body := i + len("/*")
	end := strings.Index(sql[body:], "*/")
	if end == -1 {
		return len(sql)
	}

	return body + end + len("*/")
}
