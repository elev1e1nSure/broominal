<div align="center">

# Broominal

**Безопасная, прозрачная, отменяемая очистка Windows из терминала.**

[![Go](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](https://go.dev)
[![CI](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml/badge.svg)](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/elev1e1nSure/broominal?label=release)](https://github.com/elev1e1nSure/broominal/releases)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?logo=windows)](https://github.com/elev1e1nSure/broominal)

[🇺🇸 English](README.md)

</div>

---

## Описание

Broominal — это **CLI/TUI-утилита для очистки Windows**, построенная вокруг одного правила: очистка должна быть обратимой. Она перемещает выбранные файлы в локальный **карантин** с JSON-манифестами и отчётами, чтобы каждую очистку можно было проверить и восстановить.

Это скучная и предсказуемая очистка — без фейковой магии “ускорения ПК”.

- **Безопасно по умолчанию**: карантин, а не безвозвратное удаление
- **Прозрачно**: результаты сканирования, отчёты и манифесты в обычном JSON
- **Отменяемо**: восстановление по ID очистки или последней очистки
- **Интерактивно**: TUI на Bubbletea для выбора категорий, предпросмотра и восстановления
- **Мультиязычно**: английский и русский с автоопределением при первом запуске

---

## Возможности

| Возможность | Описание |
|------------|----------|
| 🧹 **Умное сканирование** | Temp, Downloads, кэш браузеров, Корзина, логи, старые установщики, большие старые файлы, миниатюры, DirectX Shader Cache, Delivery Optimization, Windows Error Reports, кэш Discord, кэш Steam, кэш VSCode, Edge Code Cache, Chrome Code Cache, Firefox Cache2, старые Temp, старые .tmp/.log/.bak, пустые папки, npm Cache, pip Cache, Windows Update Cache, Crash & Memory Dumps, Nvidia Installer Leftovers, Telegram Desktop Cache |
| 🛡️ **Уровни риска** | `safe` / `review` / `danger` — системные пути и защищённые расширения никогда не чистятся автоматически |
| 🔄 **Отменяемая очистка** | У каждой очистки есть ID восстановления; `restore <id>` возвращает файлы обратно |
| ⚡ **Dry-Run** | `--dry-run` в CLI и `T` в TUI показывают, что произойдёт, без перемещения файлов |
| ⚙️ **Настройки** | JSON-конфиг для порогов, исключений, включения категорий, переопределений риска и языка |
| 🩺 **Doctor** | Проверки прав, директорий, манифестов и статистики карантина |
| 🗑️ **Очистка карантина** | Удаление старых карантинов с предпросмотром `--dry-run` и явным подтверждением |
| 🖥️ **TUI** | Интерактивный интерфейс на Bubbletea: главное меню, сканирование, восстановление, doctor, просмотр конфига, очистка карантина и выбор языка |
| 🌐 **i18n** | Английский / русский. Автоопределение языка по IP при первом запуске |

---

## Модель безопасности

Broominal разделяет цели очистки по риску:

| Риск | Поведение по умолчанию | Примеры |
|------|------------------------|---------|
| `safe` | выбрано по умолчанию | Temp, миниатюры, shader cache, обычные кэши приложений |
| `review` | пользователь выбирает вручную | Downloads, дампы, Windows Update cache, Telegram cache |
| `danger` | никогда не чистится автоматически | системные пути, защищённые расширения, неизвестные рискованные места |

Файлы перемещаются в:

```text
%LOCALAPPDATA%\broominal\quarantine\<restore-id>
```

Каждая очистка сохраняет `manifest.json`, где записано соответствие исходных путей и путей в карантине для восстановления.

---

## Быстрый старт

### Установка

```powershell
# Из исходников (требуется Go 1.26.3+)
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest
```

Или скачайте последнюю `.exe` из [Releases](../../releases).

### Использование

```powershell
# Сканировать безопасные зоны
broominal scan

# Запустить интерактивный TUI
broominal ui

# Очистить только безопасные элементы
broominal clean --safe

# Симулировать очистку без перемещения файлов
broominal clean --dry-run

# Восстановить конкретную очистку
broominal restore <id>

# Восстановить с перезаписью существующих файлов
broominal restore <id> --force-overwrite

# Запустить проверки здоровья
broominal doctor

# Показать конфиг
broominal config

# Удалить карантины старше 30 дней
broominal quarantine-cleanup --dry-run
broominal quarantine-cleanup --force
broominal quarantine-cleanup --force --max-age-days 7
```

---

## Сборка из исходников

Требуется **Go 1.26.3+**.

```powershell
# Клонировать и собрать
go build -o broominal.exe ./cmd/broominal

# Запустить TUI
.\broominal.exe ui
```

---

## Архитектура

```text
cmd/broominal/      Точка входа CLI (Cobra)
pkg/
  scanner/          Поиск файлов по категориям очистки
  cleaner/          Пайплайн перемещения в карантин и сохранения отчёта
  quarantine/       Перемещение / восстановление / очистка с JSON-манифестами
  report/           Генерация JSON-отчётов
  risk/             Классификация риска по путям, расширениям и конфигу
  config/           JSON-конфигурация и значения по умолчанию
  doctor/           Проверки состояния окружения
  i18n/             EN/RU-локализация и определение языка
  style/            ANSI-цвета для CLI-вывода
  util/             Форматирование размеров и общие helpers
  types/            Общие доменные типы
internal/
  tui/              Интерактивный интерфейс на Bubbletea
```

---

## Разработка

### Githooks

Включить общие хуки для проверки стиля и формата коммитов:

```powershell
git config core.hooksPath githooks
```

Хуки:

- `pre-commit` — предупреждает, если изменения кода могут требовать обновления документации
- `commit-msg` — проверяет [Conventional Commits](https://www.conventionalcommits.org/) (`feat|fix|chore|refactor|docs|test|build|ci|perf|style|revert`)

### CI / CD

Все push и pull request в `main` запускают:

- проверку `gofmt`
- `go vet`
- `golangci-lint`
- `go test ./...`
- сборку Windows-артефакта

### Релиз

Запустите workflow **Release** в GitHub Actions. Он:

1. Генерирует release notes из Conventional Commits через `git-cliff`
2. Собирает `broominal.exe`
3. Создаёт подписанный тег и GitHub Release с `checksums.txt`

---

## Участие

Баг-репорты, идеи новых категорий очистки, улучшения безопасности и Windows edge cases приветствуются. См. [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Лицензия

[MIT](LICENSE) © elev1e1nSure
