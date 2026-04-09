package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/etozhealkhipce/git-ai-summary/internal/ai"
	"github.com/etozhealkhipce/git-ai-summary/internal/git"
	"github.com/etozhealkhipce/git-ai-summary/internal/model"
	"github.com/etozhealkhipce/git-ai-summary/internal/prompt"
	"github.com/etozhealkhipce/git-ai-summary/internal/render"
	"github.com/etozhealkhipce/git-ai-summary/internal/setup"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "setup" {
		if err := setup.Run(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "git-ai-summary: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "git-ai-summary: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("git-ai-summary", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	repo := fs.String("repo", "", "path to git repo (default: current directory)")
	since := fs.String("since", "7 days ago", "git log --since value")
	maxCommits := fs.Int("max-commits", 200, "max commits to include in logs")
	maxChars := fs.Int("max-chars", 120000, "truncate bundle to this many characters (0 = no limit)")
	withStat := fs.Bool("with-stat", false, "append git diff --stat for since..HEAD")

	provider := fs.String("provider", "", "openai | anthropic | openai-compatible (default: GIT_AI_SUMMARY_PROVIDER or openai)")
	modelName := fs.String("model", "", "model id (default: GIT_AI_SUMMARY_MODEL or per provider)")
	baseURL := fs.String("base-url", "", "for openai-compatible: API base (default: GIT_AI_SUMMARY_BASE_URL)")
	apiKey := fs.String("api-key", "", "API key (overrides env)")

	timeoutSec := fs.Int("timeout", 120, "HTTP timeout seconds")
	outPath := fs.String("o", "", "write output to file (UTF-8)")
	format := fs.String("format", "tsv", "tsv | csv | md | json")
	lang := fs.String("language", "ru", "ru | en (prompt hint)")
	dryRun := fs.Bool("dry-run", false, "print git bundle only, do not call API")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	gb, err := git.Collect(git.Options{
		Repo:       *repo,
		Since:      *since,
		MaxCommits: *maxCommits,
		MaxChars:   *maxChars,
		WithStat:   *withStat,
	})
	if err != nil {
		return err
	}

	if *dryRun {
		_, _ = fmt.Fprint(os.Stdout, gb.Body)
		return nil
	}

	prov := strings.TrimSpace(*provider)
	if prov == "" {
		prov = strings.TrimSpace(os.Getenv("GIT_AI_SUMMARY_PROVIDER"))
	}
	if prov == "" {
		prov = "openai"
	}
	prov = strings.ToLower(prov)

	bu := strings.TrimSpace(*baseURL)
	if bu == "" {
		bu = strings.TrimSpace(os.Getenv("GIT_AI_SUMMARY_BASE_URL"))
	}

	key := strings.TrimSpace(*apiKey)
	if key == "" {
		switch prov {
		case "openai", "openai-compatible":
			key = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
		case "anthropic":
			key = strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
		}
	}
	if key == "" {
		return fmt.Errorf("missing API key: set -api-key or %s", keyHint(prov))
	}

	m := strings.TrimSpace(*modelName)
	if m == "" {
		m = strings.TrimSpace(os.Getenv("GIT_AI_SUMMARY_MODEL"))
	}
	if m == "" {
		switch prov {
		case "openai", "openai-compatible":
			m = "gpt-4o-mini"
		case "anthropic":
			m = "claude-3-5-sonnet-20241022"
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSec)*time.Second)
	defer cancel()

	sys, user := prompt.Messages(*lang, gb.Body)

	var completer ai.Completer
	httpClient := ai.DefaultHTTPClient(time.Duration(*timeoutSec) * time.Second)

	switch prov {
	case "openai":
		completer = &ai.OpenAIClient{
			BaseURL:    "https://api.openai.com/v1",
			APIKey:     key,
			Model:      m,
			HTTPClient: httpClient,
		}
	case "openai-compatible":
		if bu == "" {
			return fmt.Errorf("openai-compatible requires -base-url or GIT_AI_SUMMARY_BASE_URL")
		}
		completer = &ai.OpenAIClient{
			BaseURL:    bu,
			APIKey:     key,
			Model:      m,
			HTTPClient: httpClient,
		}
	case "anthropic":
		completer = &ai.AnthropicClient{
			APIKey:     key,
			Model:      m,
			HTTPClient: httpClient,
		}
	default:
		return fmt.Errorf("unknown provider %q", prov)
	}

	raw, err := completer.Complete(ctx, sys, user)
	if err != nil {
		return err
	}

	jsonStr := model.ExtractJSONObject(raw)
	sr, err := model.ParseAndValidateJSON(jsonStr)
	if err != nil {
		return fmt.Errorf("model output invalid: %w\nraw (truncated): %s", err, truncateStr(raw, 2000))
	}

	f := render.Format(strings.ToLower(strings.TrimSpace(*format)))
	out, err := render.Render(f, sr)
	if err != nil {
		return err
	}

	if strings.TrimSpace(*outPath) != "" {
		return os.WriteFile(*outPath, []byte(out), 0o644)
	}
	_, err = fmt.Print(out)
	return err
}

func keyHint(p string) string {
	switch strings.ToLower(p) {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	default:
		return "OPENAI_API_KEY"
	}
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
