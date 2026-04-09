# git-ai-summary

CLI: собирает контекст из **git** (логи и список файлов за период), отправляет в **OpenAI / Anthropic / OpenAI-compatible** API и печатает таблицу **TSV / CSV / Markdown / JSON** для демо или отчётов.

## Требования

- **Пребилды:** только `git` в `PATH` (и при установке через скрипт — `curl`).
- **Сборка из исходников:** [Go](https://go.dev/dl/) 1.22+ и `git`.

## Установка

Поддерживаются **macOS** (Apple Silicon, Intel), **Linux** (x86_64, aarch64), **Windows** (x86_64, ARM64). Готовые бинарники публикуются в [Releases](https://github.com/etozhealkhipce/git-ai-summary/releases). Сборки Linux — статические (`CGO_ENABLED=0`); на очень старых дистрибутивах при проблемах используйте установку из исходников.

### macOS и Linux

Последний релиз:

```bash
curl -sSL https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.sh | sh
```

Зафиксировать версию:

```bash
GIT_AI_SUMMARY_VERSION=0.1.0 curl -sSL https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.sh | sh
```

Переменные установки:

- **GIT_AI_SUMMARY_VERSION** — версия релиза без префикса `v` (например `0.1.0`).
- **GIT_AI_SUMMARY_INSTALL_DIR** — каталог для бинарника (по умолчанию `$HOME/.local/bin`).

Безопаснее, чем `curl | sh`: склонируйте репозиторий и выполните `./install.sh` локально после просмотра скрипта.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.ps1 | iex
```

С версией и своим каталогом:

```powershell
$env:GIT_AI_SUMMARY_VERSION = "0.1.0"
$env:GIT_AI_SUMMARY_INSTALL_DIR = "C:\tools\git-ai-summary"
irm https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.ps1 | iex
```

По умолчанию бинарник в `%LOCALAPPDATA%\git-ai-summary`. Добавьте каталог в **PATH** пользователя, если установщик напомнит об этом.

### Релизы для поддержки (maintainers)

Релиз выполняется **только вручную из GitHub**: откройте [Actions](https://github.com/etozhealkhipce/git-ai-summary/actions) → workflow **release** → **Run workflow**. Выберите ветку (обычно `main`) и тип бампа:

- **auto** — следующая версия по [Conventional Commits](https://www.conventionalcommits.org/) с последнего тега (`fix:` → patch, `feat:` → minor, `BREAKING CHANGE` / `feat!:` / `fix!:` → major, `chore:` и т.п. без релизного бампа не увеличивают версию).
- **patch** / **minor** / **major** — принудительно увеличить соответствующую часть semver относительно последнего тега.

Workflow сам посчитает версию ([svu](https://github.com/caarlos0/svu)), создаст аннотированный тег `v*` на текущем коммите и запустит [GoReleaser](https://goreleaser.com/) ([.github/workflows/release.yml](.github/workflows/release.yml)).

Если в репозитории ещё **нет ни одного тега**, первый запуск зафиксирует релиз как **v0.1.0** (независимо от выбранного `bump`).

### Настройка API после установки

Интерактивно (ключи и переменные в профиль оболочки или файл для PowerShell):

```bash
git-ai-summary setup
```

### Установка из исходников

```bash
go install github.com/etozhealkhipce/git-ai-summary/cmd/git-ai-summary@latest
```

Бинарник в **`$GOBIN`** или **`$GOPATH/bin`** — каталог должен быть в **`PATH`**.

Локальная сборка:

```bash
git clone https://github.com/etozhealkhipce/git-ai-summary.git
cd git-ai-summary
go build -buildvcs=false -o git-ai-summary ./cmd/git-ai-summary
```

## Переменные окружения

| Переменная | Назначение |
|------------|------------|
| `OPENAI_API_KEY` | Ключ для `openai` и `openai-compatible`. |
| `ANTHROPIC_API_KEY` | Ключ для `anthropic`. |
| `GIT_AI_SUMMARY_PROVIDER` | `openai`, `anthropic` или `openai-compatible`, если не указан `-provider`. |
| `GIT_AI_SUMMARY_BASE_URL` | Базовый URL для `openai-compatible`, если не указан `-base-url`. |
| `GIT_AI_SUMMARY_MODEL` | ID модели, если не указан `-model`. |

Флаги командной строки имеют приоритет над переменными окружения.

## Использование

```text
git-ai-summary [-repo path] [-since "7 days ago"] [-max-commits 200] [-max-chars 120000]
  [-with-stat] [-provider openai|anthropic|openai-compatible] [-model ...] [-base-url ...]
  [-api-key ...] [-timeout 120] [-format tsv|csv|md|json] [-o file] [-language ru|en]
  [-dry-run]

git-ai-summary setup
```

- По умолчанию репозиторий — **текущая директория** (`-repo` можно не указывать).
- **`-provider openai-compatible`** требует **`-base-url`** или **`GIT_AI_SUMMARY_BASE_URL`** (например `https://api.openai.com/v1` или базу совместимого провайдера с суффиксом `/v1`).

### Примеры

Сборка контекста без вызова API:

```bash
cd ~/work/my-app
git-ai-summary -dry-run -since "2 weeks ago"
```

OpenAI, TSV в файл:

```bash
export OPENAI_API_KEY=...
git-ai-summary -format tsv -o summary.tsv -since "1 week ago"
```

Anthropic:

```bash
export ANTHROPIC_API_KEY=...
git-ai-summary -provider anthropic -format md -o summary.md
```

## Безопасность

Не коммитьте ключи. Для CI используйте секреты окружения. Команда `setup` может записать ключи в ваш профиль оболочки — это осознанный компромисс удобства; при необходимости задавайте переменные только в сессии или в секрет-хранилище.

## Лицензия

MIT, см. [LICENSE](LICENSE).
