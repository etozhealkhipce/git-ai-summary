package prompt

import (
	"fmt"
	"strings"
)

// Messages returns system and user content for the LLM.
func Messages(language, gitBundle string) (system, user string) {
	langHint := "Write human-readable text in Russian."
	if strings.EqualFold(language, "en") {
		langHint = "Write human-readable text in English."
	}

	system = strings.TrimSpace(fmt.Sprintf(`
You are a release-notes assistant. Given git metadata and changed files, produce a demo-friendly summary.

Rules:
- Return ONLY a single JSON object. No markdown fences, no commentary before or after.
- JSON shape (all keys required except optional fields):
  {"rows":[{"area":"","path_or_url":"","summary":"","ticket":""}],"notes":[]}
  - path_or_url: route/URL if inferable (e.g. /main, /api/docs), else file path or "—"
  - summary: one short line: what changed and why it matters for a demo
  - ticket: issue id from commit messages if present (e.g. SQBC-712), else omit or ""
  - area: optional grouping (e.g. merchant, auth, api)
  - notes: optional global bullets (strings)
- Infer routes only when reasonable from paths (routes.tsx, paths constants); do not invent URLs.
- %s
`, langHint))

	user = "Git context:\n\n" + gitBundle
	return system, user
}
