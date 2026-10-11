package common

import (
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// rtfSkipDestinations are the RTF groups that hold no text of the document (the tables of the fonts,
// colors and styles, the info and the pictures).
var rtfSkipDestinations = map[string]bool{
	"fonttbl": true, "colortbl": true, "stylesheet": true, "info": true, "pict": true,
	"generator": true, "listtable": true, "listoverridetable": true, "rsidtbl": true, "themedata": true,
}

// RtfToText returns the plain text of an RTF value (e.g. the napomena of a partner the legacy saved
// from a rich edit control): the paragraphs as new lines, the characters \'hh in the code page 1250 and
// \uN as Unicode, without the control words and the groups of the fonts, colors and styles. A value that
// is not RTF is returned as it is.
func RtfToText(value string) string {
	if !strings.HasPrefix(strings.TrimSpace(value), `{\rtf`) {
		return value
	}
	var out strings.Builder
	// skip of every open group: a group is skipped when its parent is or when it is a destination
	// without text ({\* ...} or one of rtfSkipDestinations).
	skip := []bool{false}
	groupStart := false
	ucSkip := 0 // the characters after \uN to drop (the ANSI substitute of the Unicode character)
	write := func(text string) {
		if !skip[len(skip)-1] {
			out.WriteString(text)
		}
	}
	for i := 0; i < len(value); i++ {
		ch := value[i]
		switch ch {
		case '{':
			skip = append(skip, skip[len(skip)-1])
			groupStart = true
			continue
		case '}':
			if len(skip) > 1 {
				skip = skip[:len(skip)-1]
			}
			groupStart = false
			continue
		case '\r', '\n':
			continue
		case '\\':
			if i+1 >= len(value) {
				continue
			}
			next := value[i+1]
			switch {
			case next == '\\' || next == '{' || next == '}':
				write(string(next))
				i++
			case next == '*':
				skip[len(skip)-1] = true
				i++
			case next == '\'':
				if i+3 < len(value) {
					if b, err := strconv.ParseUint(value[i+2:i+4], 16, 8); err == nil {
						if ucSkip > 0 {
							ucSkip--
						} else {
							write(string(charmap.Windows1250.DecodeByte(byte(b))))
						}
					}
				}
				i += 3
			case (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z'):
				j := i + 1
				for j < len(value) && ((value[j] >= 'a' && value[j] <= 'z') || (value[j] >= 'A' && value[j] <= 'Z')) {
					j++
				}
				word := value[i+1 : j]
				k := j
				if k < len(value) && value[k] == '-' {
					k++
				}
				for k < len(value) && value[k] >= '0' && value[k] <= '9' {
					k++
				}
				param := value[j:k]
				if k < len(value) && value[k] == ' ' {
					k++
				}
				i = k - 1
				if groupStart && rtfSkipDestinations[word] {
					skip[len(skip)-1] = true
				}
				switch word {
				case "par", "line":
					write("\n")
				case "tab":
					write("\t")
				case "u":
					if n, err := strconv.Atoi(param); err == nil {
						if n < 0 {
							n += 65536
						}
						write(string(rune(n)))
						ucSkip = 1
					}
				}
			default:
				i++
			}
			groupStart = false
			continue
		}
		groupStart = false
		if ucSkip > 0 {
			ucSkip--
			continue
		}
		write(string(ch))
	}
	return strings.TrimSpace(out.String())
}
