package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/etozhealkhipce/git-ai-summary/internal/model"
)

// Format is output format.
type Format string

const (
	FormatTSV  Format = "tsv"
	FormatCSV  Format = "csv"
	FormatMD   Format = "md"
	FormatJSON Format = "json"
)

func Render(f Format, sr *model.SummaryResponse) (string, error) {
	switch f {
	case FormatTSV:
		return renderTSV(sr), nil
	case FormatCSV:
		return renderCSV(sr), nil
	case FormatMD:
		return renderMD(sr), nil
	case FormatJSON:
		b, err := json.MarshalIndent(sr, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b) + "\n", nil
	default:
		return "", fmt.Errorf("unknown format %q", f)
	}
}

func normalizeCell(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\t", " "), "\n", " ")
}

func renderTSV(sr *model.SummaryResponse) string {
	var b strings.Builder
	b.WriteString("area\tpath_or_url\tsummary\tticket\n")
	for _, r := range sr.Rows {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\n",
			normalizeCell(r.Area),
			normalizeCell(r.PathOrURL),
			normalizeCell(r.Summary),
			normalizeCell(r.Ticket),
		)
	}
	if len(sr.Notes) > 0 {
		b.WriteString("\nnotes\n")
		for _, n := range sr.Notes {
			fmt.Fprintf(&b, "- %s\n", normalizeCell(n))
		}
	}
	return b.String()
}

func escapeCSVField(s string) string {
	if strings.ContainsAny(s, `",`+"\n\r") {
		s = strings.ReplaceAll(s, `"`, `""`)
		return `"` + s + `"`
	}
	return s
}

func renderCSV(sr *model.SummaryResponse) string {
	var b strings.Builder
	b.WriteString("area,path_or_url,summary,ticket\n")
	for _, r := range sr.Rows {
		fmt.Fprintf(&b, "%s,%s,%s,%s\n",
			escapeCSVField(r.Area),
			escapeCSVField(r.PathOrURL),
			escapeCSVField(r.Summary),
			escapeCSVField(r.Ticket),
		)
	}
	if len(sr.Notes) > 0 {
		b.WriteString("\nnotes\n")
		for _, n := range sr.Notes {
			fmt.Fprintf(&b, "%s\n", escapeCSVField(n))
		}
	}
	return b.String()
}

func renderMD(sr *model.SummaryResponse) string {
	var b strings.Builder
	b.WriteString("| area | path_or_url | summary | ticket |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, r := range sr.Rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
			mdCell(r.Area), mdCell(r.PathOrURL), mdCell(r.Summary), mdCell(r.Ticket))
	}
	if len(sr.Notes) > 0 {
		b.WriteString("\n**notes**\n\n")
		for _, n := range sr.Notes {
			fmt.Fprintf(&b, "- %s\n", mdCell(n))
		}
	}
	return b.String()
}

func mdCell(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
