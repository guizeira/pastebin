package main

import (
	"html"
	"strconv"
	"strings"
)

const ansiESC = '\x1b'

var ansiFG = map[int]string{
	30: "#000000", 31: "#ef4444", 32: "#22c55e", 33: "#eab308",
	34: "#3b82f6", 35: "#d946ef", 36: "#06b6d4", 37: "#e4e4e7",
	90: "#71717a", 91: "#f87171", 92: "#4ade80", 93: "#facc15",
	94: "#60a5fa", 95: "#e879f9", 96: "#22d3ee", 97: "#fafafa",
}

var ansiBG = map[int]string{
	40: "#000000", 41: "#7f1d1d", 42: "#14532d", 43: "#713f12",
	44: "#1e3a8a", 45: "#701a75", 46: "#164e63", 47: "#a1a1aa",
	100: "#3f3f46", 101: "#991b1b", 102: "#166534", 103: "#a16207",
	104: "#1d4ed8", 105: "#a21caf", 106: "#0e7490", 107: "#e4e4e7",
}

var ansi256Basic = []string{
	"#000000", "#ef4444", "#22c55e", "#eab308",
	"#3b82f6", "#d946ef", "#06b6d4", "#e4e4e7",
	"#71717a", "#f87171", "#4ade80", "#facc15",
	"#60a5fa", "#e879f9", "#22d3ee", "#fafafa",
}

type ansiStyle struct {
	bold, dim, underline bool
	fg, bg               string
}

func hasANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

func ansi256Color(n int) string {
	if n < 0 || n > 255 {
		return ""
	}
	if n < 16 {
		return ansi256Basic[n]
	}
	if n < 232 {
		n -= 16
		r := n / 36
		g := (n % 36) / 6
		b := n % 6
		steps := []int{0, 95, 135, 175, 215, 255}
		return "rgb(" + strconv.Itoa(steps[r]) + "," + strconv.Itoa(steps[g]) + "," + strconv.Itoa(steps[b]) + ")"
	}
	gray := 8 + (n-232)*10
	return "rgb(" + strconv.Itoa(gray) + "," + strconv.Itoa(gray) + "," + strconv.Itoa(gray) + ")"
}

func (s *ansiStyle) css() string {
	var parts []string
	if s.bold {
		parts = append(parts, "font-weight:700")
	}
	if s.dim {
		parts = append(parts, "opacity:0.65")
	}
	if s.underline {
		parts = append(parts, "text-decoration:underline")
	}
	if s.fg != "" {
		parts = append(parts, "color:"+s.fg)
	}
	if s.bg != "" {
		parts = append(parts, "background-color:"+s.bg)
	}
	return strings.Join(parts, ";")
}

func (s *ansiStyle) reset() {
	*s = ansiStyle{}
}

func (s *ansiStyle) apply(codes []string) {
	if len(codes) == 0 || (len(codes) == 1 && codes[0] == "") {
		codes = []string{"0"}
	}
	for i := 0; i < len(codes); i++ {
		c, err := strconv.Atoi(codes[i])
		if err != nil {
			s.reset()
			continue
		}
		switch {
		case c == 0:
			s.reset()
		case c == 1:
			s.bold = true
		case c == 2:
			s.dim = true
		case c == 4:
			s.underline = true
		case c == 22:
			s.bold = false
			s.dim = false
		case c == 24:
			s.underline = false
		case c == 39:
			s.fg = ""
		case c == 49:
			s.bg = ""
		case ansiFG[c] != "":
			s.fg = ansiFG[c]
		case ansiBG[c] != "":
			s.bg = ansiBG[c]
		case c == 38 || c == 48:
			isFG := c == 38
			if i+1 >= len(codes) {
				continue
			}
			i++
			mode, _ := strconv.Atoi(codes[i])
			if mode == 5 && i+1 < len(codes) {
				i++
				n, _ := strconv.Atoi(codes[i])
				color := ansi256Color(n)
				if isFG {
					s.fg = color
				} else {
					s.bg = color
				}
			} else if mode == 2 && i+3 < len(codes) {
				r, _ := strconv.Atoi(codes[i+1])
				g, _ := strconv.Atoi(codes[i+2])
				b, _ := strconv.Atoi(codes[i+3])
				i += 3
				rgb := "rgb(" + strconv.Itoa(r) + "," + strconv.Itoa(g) + "," + strconv.Itoa(b) + ")"
				if isFG {
					s.fg = rgb
				} else {
					s.bg = rgb
				}
			}
		}
	}
}

func appendStyled(b *strings.Builder, text string, style *ansiStyle) {
	if text == "" {
		return
	}
	escaped := html.EscapeString(text)
	if css := style.css(); css != "" {
		b.WriteString(`<span style="`)
		b.WriteString(css)
		b.WriteString(`">`)
		b.WriteString(escaped)
		b.WriteString(`</span>`)
		return
	}
	b.WriteString(escaped)
}

// ansiToHTML converts ANSI SGR sequences into safe colored HTML spans.
func ansiToHTML(input string) string {
	var b strings.Builder
	b.Grow(len(input) + 64)
	style := ansiStyle{}

	for len(input) > 0 {
		i := strings.IndexByte(input, ansiESC)
		if i < 0 {
			appendStyled(&b, input, &style)
			break
		}
		if i > 0 {
			appendStyled(&b, input[:i], &style)
			input = input[i:]
		}
		// ESC[
		if len(input) < 2 || input[1] != '[' {
			appendStyled(&b, input[:1], &style)
			input = input[1:]
			continue
		}
		end := strings.IndexByte(input, 'm')
		if end < 0 {
			appendStyled(&b, input, &style)
			break
		}
		params := input[2:end]
		style.apply(strings.Split(params, ";"))
		input = input[end+1:]
	}
	return b.String()
}
