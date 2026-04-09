package setup

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	markerBegin = "# git-ai-summary"
	markerEnd   = "# end git-ai-summary"
)

// Run interactive configuration for shell or Windows env file.
func Run(args []string) error {
	_ = args // reserved for future flags (e.g. -y)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("setup requires an interactive terminal")
	}

	scan := bufio.NewScanner(os.Stdin)
	lang := askLang(scan)
	prov := askProvider(scan, lang)
	baseURL := ""
	if prov == "openai-compatible" {
		baseURL = askBaseURL(scan, lang)
	}
	key := askAPIKey(lang)
	if key == "" {
		return fmt.Errorf("%s", msgNoKey(lang))
	}
	model := askModel(scan, lang)
	if askVerify(scan, lang) {
		if prov == "anthropic" {
			fmt.Fprintf(os.Stdout, "%s\n", skipAnthropicVerify(lang))
		} else if err := verifyKey(prov, baseURL, key); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", warnLabel(lang), err)
		} else {
			fmt.Fprintf(os.Stdout, "%s\n", okVerify(lang))
		}
	}

	if runtime.GOOS == "windows" {
		return writeWindows(buildPSBlock(prov, baseURL, key, model), lang, scan)
	}
	return writeUnix(buildBlock(prov, baseURL, key, model), lang, scan)
}

func askLang(scan *bufio.Scanner) string {
	fmt.Println("Language / язык: [1] English  [2] Russian (default 2)")
	if !scan.Scan() {
		return "ru"
	}
	switch strings.TrimSpace(scan.Text()) {
	case "1", "en", "EN":
		return "en"
	default:
		return "ru"
	}
}

func askProvider(scan *bufio.Scanner, lang string) string {
	if lang == "en" {
		fmt.Println("Provider: [1] openai  [2] anthropic  [3] openai-compatible (default 1)")
	} else {
		fmt.Println("Провайдер: [1] openai  [2] anthropic  [3] openai-compatible (по умолчанию 1)")
	}
	if !scan.Scan() {
		return "openai"
	}
	switch strings.TrimSpace(scan.Text()) {
	case "2":
		return "anthropic"
	case "3":
		return "openai-compatible"
	default:
		return "openai"
	}
}

func askBaseURL(scan *bufio.Scanner, lang string) string {
	if lang == "en" {
		fmt.Print("Base URL (e.g. https://api.openai.com/v1): ")
	} else {
		fmt.Print("Базовый URL (например https://api.openai.com/v1): ")
	}
	if !scan.Scan() {
		return ""
	}
	return strings.TrimSpace(scan.Text())
}

func askAPIKey(lang string) string {
	if lang == "en" {
		fmt.Print("API key (input hidden): ")
	} else {
		fmt.Print("API-ключ (ввод скрыт): ")
	}
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func askModel(scan *bufio.Scanner, lang string) string {
	if lang == "en" {
		fmt.Print("Model ID (Enter = default for provider): ")
	} else {
		fmt.Print("ID модели (Enter = значение по умолчанию): ")
	}
	if !scan.Scan() {
		return ""
	}
	return strings.TrimSpace(scan.Text())
}

func askVerify(scan *bufio.Scanner, lang string) bool {
	if lang == "en" {
		fmt.Print("Test API key with a quick HTTP request? [y/N]: ")
	} else {
		fmt.Print("Проверить ключ простым HTTP-запросом? [y/N]: ")
	}
	if !scan.Scan() {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(scan.Text()))
	return s == "y" || s == "yes" || s == "д" || s == "да"
}

func buildBlock(prov, baseURL, key, model string) string {
	var b strings.Builder
	b.WriteString(markerBegin + "\n")
	b.WriteString(fmt.Sprintf("export GIT_AI_SUMMARY_PROVIDER=%q\n", prov))
	switch prov {
	case "anthropic":
		b.WriteString(fmt.Sprintf("export ANTHROPIC_API_KEY=%q\n", key))
	default:
		b.WriteString(fmt.Sprintf("export OPENAI_API_KEY=%q\n", key))
	}
	if prov == "openai-compatible" && baseURL != "" {
		b.WriteString(fmt.Sprintf("export GIT_AI_SUMMARY_BASE_URL=%q\n", baseURL))
	}
	if model != "" {
		b.WriteString(fmt.Sprintf("export GIT_AI_SUMMARY_MODEL=%q\n", model))
	}
	b.WriteString(markerEnd + "\n")
	return b.String()
}

func buildPSBlock(prov, baseURL, key, model string) string {
	var b strings.Builder
	b.WriteString("# git-ai-summary — dot-source in PowerShell: . $HOME\\.git-ai-summary-env.ps1\n")
	b.WriteString(fmt.Sprintf("$env:GIT_AI_SUMMARY_PROVIDER = %q\n", prov))
	switch prov {
	case "anthropic":
		b.WriteString(fmt.Sprintf("$env:ANTHROPIC_API_KEY = %q\n", key))
	default:
		b.WriteString(fmt.Sprintf("$env:OPENAI_API_KEY = %q\n", key))
	}
	if prov == "openai-compatible" && baseURL != "" {
		b.WriteString(fmt.Sprintf("$env:GIT_AI_SUMMARY_BASE_URL = %q\n", baseURL))
	}
	if model != "" {
		b.WriteString(fmt.Sprintf("$env:GIT_AI_SUMMARY_MODEL = %q\n", model))
	}
	return b.String()
}

func writeUnix(bl, lang string, scan *bufio.Scanner) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rc := filepath.Join(home, ".zshrc")
	if sh := os.Getenv("SHELL"); !strings.Contains(sh, "zsh") {
		rc = filepath.Join(home, ".bashrc")
	}
	if lang == "en" {
		fmt.Printf("Append to %s? [Y/n]: ", rc)
	} else {
		fmt.Printf("Добавить в %s? [Y/n]: ", rc)
	}
	if scan.Scan() {
		s := strings.ToLower(strings.TrimSpace(scan.Text()))
		if s == "n" || s == "no" || s == "н" {
			fmt.Println(bl)
			return nil
		}
	}
	data, err := os.ReadFile(rc)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	s := string(data)
	s = stripMarked(s)
	s = strings.TrimRight(s, "\n") + "\n\n" + bl + "\n"
	if err := os.WriteFile(rc, []byte(s), 0o644); err != nil {
		return err
	}
	if lang == "en" {
		fmt.Fprintf(os.Stdout, "Wrote configuration to %s. Open a new terminal or run: source %s\n", rc, rc)
	} else {
		fmt.Fprintf(os.Stdout, "Настройки записаны в %s. Откройте новый терминал или выполните: source %s\n", rc, rc)
	}
	return nil
}

func writeWindows(blPS, lang string, scan *bufio.Scanner) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".git-ai-summary-env.ps1")
	if lang == "en" {
		fmt.Printf("Write %s and load it from your PowerShell profile? [Y/n]: ", path)
	} else {
		fmt.Printf("Записать %s и подключать из профиля PowerShell? [Y/n]: ", path)
	}
	if scan.Scan() {
		s := strings.ToLower(strings.TrimSpace(scan.Text()))
		if s == "n" || s == "no" || s == "н" {
			fmt.Println(blPS)
			return nil
		}
	}
	if err := os.WriteFile(path, []byte(blPS+"\n"), 0o644); err != nil {
		return err
	}
	loadLine := fmt.Sprintf(". '%s'\n", path)

	profile := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	if err := ensureProfileLine(profile, loadLine); err != nil {
		if lang == "en" {
			fmt.Fprintf(os.Stderr, "Wrote %s. Add this line to your PowerShell profile:\n. '%s'\n", path, path)
		} else {
			fmt.Fprintf(os.Stderr, "Создан файл %s. Добавьте в профиль PowerShell строку:\n. '%s'\n", path, path)
		}
		return nil
	}
	if lang == "en" {
		fmt.Fprintf(os.Stdout, "Wrote %s and updated %s. Restart PowerShell.\n", path, profile)
	} else {
		fmt.Fprintf(os.Stdout, "Записано %s, обновлён %s. Перезапустите PowerShell.\n", path, profile)
	}
	return nil
}

func ensureProfileLine(profile, line string) error {
	if err := os.MkdirAll(filepath.Dir(profile), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(profile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	s := string(data)
	if strings.Contains(s, ".git-ai-summary-env.ps1") {
		return nil
	}
	s = strings.TrimRight(s, "\n") + "\n\n" + line
	return os.WriteFile(profile, []byte(s), 0o644)
}

func stripMarked(s string) string {
	start := strings.Index(s, markerBegin)
	if start < 0 {
		return s
	}
	end := strings.Index(s[start:], markerEnd)
	if end < 0 {
		return s[:start]
	}
	end += start + len(markerEnd)
	for end < len(s) && (s[end] == '\n' || s[end] == '\r') {
		end++
	}
	return s[:start] + s[end:]
}

func verifyKey(prov, baseURL, key string) error {
	client := &http.Client{Timeout: 25 * time.Second}
	switch prov {
	case "openai":
		req, err := http.NewRequest(http.MethodGet, "https://api.openai.com/v1/models", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		return doProbe(client, req)
	case "openai-compatible":
		base := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
		if base == "" {
			return errors.New("base URL missing")
		}
		req, err := http.NewRequest(http.MethodGet, base+"/models", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+key)
		return doProbe(client, req)
	default:
		return nil
	}
}

func doProbe(client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("HTTP 401: invalid API key")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func msgNoKey(lang string) string {
	if lang == "en" {
		return "empty API key"
	}
	return "пустой API-ключ"
}

func warnLabel(lang string) string {
	if lang == "en" {
		return "Warning"
	}
	return "Предупреждение"
}

func okVerify(lang string) string {
	if lang == "en" {
		return "API probe succeeded."
	}
	return "Проверка API прошла успешно."
}

func skipAnthropicVerify(lang string) string {
	if lang == "en" {
		return "Skipping HTTP probe for Anthropic (not supported)."
	}
	return "Проверка HTTP для Anthropic пропущена."
}
