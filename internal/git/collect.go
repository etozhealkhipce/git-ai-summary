package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Bundle is aggregated git context for the prompt.
type Bundle struct {
	TopLevel string
	Head     string
	Branch   string
	Remote   string
	Body     string
}

type Options struct {
	Repo       string
	Since      string
	MaxCommits int
	MaxChars   int
	WithStat   bool
}

// Collect runs git in repoDir (empty = cwd) and builds a text bundle.
func Collect(opts Options) (*Bundle, error) {
	repo := opts.Repo
	if repo == "" {
		repo = "."
	}
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		return nil, fmt.Errorf("abs repo path: %w", err)
	}

	top, err := gitOutput(repoAbs, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not a git repository (rev-parse failed): %w", err)
	}
	top = strings.TrimSpace(top)

	head, _ := gitOutput(top, "rev-parse", "--short", "HEAD")
	head = strings.TrimSpace(head)

	branch, _ := gitOutput(top, "branch", "--show-current")
	branch = strings.TrimSpace(branch)

	remote, errRem := gitOutput(top, "remote", "get-url", "origin")
	if errRem != nil {
		remote = ""
	}
	remote = strings.TrimSpace(remote)

	since := opts.Since
	if since == "" {
		since = "7 days ago"
	}
	maxN := opts.MaxCommits
	if maxN <= 0 {
		maxN = 200
	}

	var b strings.Builder
	fmt.Fprintf(&b, "repository_root: %s\n", top)
	fmt.Fprintf(&b, "branch: %s\n", branch)
	fmt.Fprintf(&b, "head: %s\n", head)
	if remote != "" {
		fmt.Fprintf(&b, "origin: %s\n", remote)
	}
	fmt.Fprintf(&b, "since: %s\n", since)
	fmt.Fprintf(&b, "\n--- commits (oneline) ---\n")

	logOneline, err := gitOutput(top,
		"log", "--since="+since, fmt.Sprintf("-n%d", maxN),
		"--pretty=format:%h %ad %s", "--date=short",
	)
	if err != nil {
		return nil, fmt.Errorf("git log oneline: %w", err)
	}
	b.WriteString(strings.TrimSpace(logOneline))
	b.WriteByte('\n')

	fmt.Fprintf(&b, "\n--- commits with name-status ---\n")
	ns, err := gitOutput(top,
		"log", "--since="+since, fmt.Sprintf("-n%d", maxN),
		"--name-status", "--pretty=format:== %h %ad %s ==",
		"--date=short",
	)
	if err != nil {
		return nil, fmt.Errorf("git log name-status: %w", err)
	}
	b.WriteString(strings.TrimSpace(ns))
	b.WriteByte('\n')

	if opts.WithStat {
		fmt.Fprintf(&b, "\n--- diff stat (since..HEAD) ---\n")
		firstHash, errF := gitOutput(top,
			"log", "--since="+since, "--reverse", "--pretty=%H", "-n", "1",
		)
		firstHash = strings.TrimSpace(firstHash)
		if errF == nil && firstHash != "" {
			stat, errSt := gitOutput(top, "diff", "--stat", firstHash+"..HEAD")
			if errSt == nil {
				b.WriteString(strings.TrimSpace(stat))
				b.WriteByte('\n')
			}
		}
	}

	body := b.String()
	if opts.MaxChars > 0 && len(body) > opts.MaxChars {
		body = body[:opts.MaxChars] + "\n\n[truncated to max-chars]\n"
	}

	return &Bundle{
		TopLevel: top,
		Head:     head,
		Branch:   branch,
		Remote:   remote,
		Body:     body,
	}, nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out.String(), nil
}
