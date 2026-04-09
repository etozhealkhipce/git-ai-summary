package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/etozhealkhipce/git-ai-summary/internal/model"
)

// PrettyOptions controls terminal-oriented rendering.
type PrettyOptions struct {
	Width  int
	Color  bool
	Locale string // "en" or default Russian labels
}

// Pretty renders a human-readable layout with word-wrapped summaries.
func Pretty(sr *model.SummaryResponse, o PrettyOptions) string {
	w := o.Width
	if w < 48 {
		w = 48
	}
	inner := w - 4
	if inner < 32 {
		inner = 32
	}

	bold := ""
	dim := ""
	cyan := ""
	reset := ""
	if o.Color {
		bold = "\x1b[1m"
		dim = "\x1b[2m"
		cyan = "\x1b[36m"
		reset = "\x1b[0m"
	}

	titleChanges := "Изменения"
	titleNotes := "Заметки"
	labelArea := "область"
	labelTicket := "тикет"
	labelSummary := "описание"
	if strings.EqualFold(o.Locale, "en") {
		titleChanges = "Changes"
		titleNotes = "Notes"
		labelArea = "area"
		labelTicket = "ticket"
		labelSummary = "summary"
	}

	var out strings.Builder
	rule := strings.Repeat("─", min(w, 76))
	fmt.Fprintf(&out, "%s%s%s\n", dim, rule, reset)
	fmt.Fprintf(&out, "%s%s%s\n\n", bold, titleChanges, reset)

	for i, r := range sr.Rows {
		if i > 0 {
			out.WriteByte('\n')
		}
		path := normalizeCell(r.PathOrURL)
		area := strings.TrimSpace(normalizeCell(r.Area))
		tick := strings.TrimSpace(normalizeCell(r.Ticket))
		sum := normalizeCell(r.Summary)
		if area == "" {
			area = "—"
		}
		if tick == "" {
			tick = "—"
		}

		fmt.Fprintf(&out, "%s%02d.%s %s%s%s%s\n", dim, i+1, reset, bold, cyan, path, reset)
		fmt.Fprintf(&out, "   %s%s:%s %s\n", dim, labelArea, reset, area)
		fmt.Fprintf(&out, "   %s%s:%s %s\n", dim, labelTicket, reset, tick)
		fmt.Fprintf(&out, "   %s%s:%s\n", dim, labelSummary, reset)
		for _, line := range wrapText(sum, inner) {
			fmt.Fprintf(&out, "      %s\n", line)
		}
	}

	if len(sr.Notes) > 0 {
		fmt.Fprintf(&out, "\n%s%s%s\n", dim, rule, reset)
		fmt.Fprintf(&out, "%s%s%s\n\n", bold, titleNotes, reset)
		for _, n := range sr.Notes {
			for _, line := range wrapText(normalizeCell(n), inner) {
				fmt.Fprintf(&out, "  %s•%s %s\n", dim, reset, line)
			}
		}
	}

	fmt.Fprintf(&out, "\n%s%s%s\n", dim, rule, reset)
	return out.String()
}

func wrapText(s string, maxWidth int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{"—"}
	}
	if maxWidth < 8 {
		maxWidth = 8
	}
	var lines []string
	words := strings.Fields(s)
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		lines = append(lines, cur.String())
		cur.Reset()
	}
	for _, word := range words {
		rw := utf8.RuneCountInString(word)
		if rw > maxWidth {
			flush()
			lines = append(lines, splitRunes(word, maxWidth)...)
			continue
		}
		sep := 0
		if cur.Len() > 0 {
			sep = 1
		}
		if utf8.RuneCountInString(cur.String())+sep+rw > maxWidth {
			flush()
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(word)
	}
	flush()
	if len(lines) == 0 {
		return []string{"—"}
	}
	return lines
}

func splitRunes(s string, maxWidth int) []string {
	var out []string
	runes := []rune(s)
	for len(runes) > 0 {
		n := min(len(runes), maxWidth)
		out = append(out, string(runes[:n]))
		runes = runes[n:]
	}
	return out
}
