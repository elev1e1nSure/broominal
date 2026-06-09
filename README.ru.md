<div align="center">

# 🧹 broominal

**безопасная, прозрачная, отменяемая очистка windows из терминала**

[![go](https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go\&logoColor=white)](https://go.dev)
[![ci](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml/badge.svg)](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/elev1e1nSure/broominal?label=release)](https://github.com/elev1e1nSure/broominal/releases)
[![license](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![platform](https://img.shields.io/badge/platform-Windows-0078D4?logo=windows\&logoColor=white)](https://github.com/elev1e1nSure/broominal)

[english](README.md) · [русский](README.ru.md)

</div>

---

## что это

**broominal** — cli/tui-утилита для очистки windows, построенная вокруг одного правила:

> очистка должна быть **обратима**

вместо безвозвратного удаления broominal перемещает выбранные файлы в локальный **карантин**, сохраняет json-манифесты и делает каждую очистку проверяемой и восстанавливаемой.

без фейкового ускорения пк и скрытых системных твиков, без очистки на честном слове.

---

## возможности

- **безопасно по умолчанию** — файлы отправляются в карантин, а не удаляются
- **прозрачно** — результаты сканирования, отчёты и манифесты — обычный json
- **отменяемо** — восстанови любую очистку по id или верни последнюю
- **предсказуемо** — явные категории, уровни риска и исключения
- **интерактивно** — tui на bubbletea для сканирования, предпросмотра и восстановления
- **мультиязычно** — русский и английский с автоопределением при первом запуске
- **25+ категорий** — temp, кэши, логи, данные браузеров, инструменты разработки и не только
- **doctor** — лёгкие проверки прав, манифестов и состояния карантина

---

## модель безопасности

> **safe** очистка выбрана по умолчанию. **review** требует ручного выбора. элементы **danger** никогда не чистятся автоматически.

```
┌─────────────────────────────────────────────────────────────┐
│  safe    ▸ выбрано по умолчанию                             │
│           temp, миниатюры, shader cache, кэши приложений    │
├─────────────────────────────────────────────────────────────┤
│  review  ▸ пользователь выбирает вручную                    │
│           downloads, дампы, windows update cache, telegram  │
├─────────────────────────────────────────────────────────────┤
│  danger  ▸ никогда не чистится автоматически                │
│           системные пути, защищённые расширения             │
└─────────────────────────────────────────────────────────────┘
```

файлы перемещаются в `%LOCALAPPDATA%\broominal\quarantine\<restore-id>` с `manifest.json`, где записано соответствие исходных путей и путей в карантине.

---

## быстрый старт

```powershell
# установка из исходников (требуется go 1.26.3+)
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest

# ...или скачай последнюю .exe из релизов
```

---

## использование

```powershell
# просканировать безопасные зоны
broominal scan

# запустить интерактивный tui
broominal ui

# очистить только безопасные элементы
broominal clean --safe

# разрешить очистку опасных элементов (требует явного подтверждения)
broominal clean --danger

# восстановить конкретную очистку
broominal restore <id>

# восстановить с перезаписью существующих файлов
broominal restore <id> --force-overwrite

# запустить проверки состояния
broominal doctor

# показать конфиг
broominal config

# предпросмотр очистки старых карантинов (покажет, что будет удалено)
broominal quarantine-cleanup

# удалить карантины старше 30 дней
broominal quarantine-cleanup --force

# удалить карантины старше N дней
broominal quarantine-cleanup --max-age-days 7 --force
```

---

## сборка из исходников

```powershell
git clone https://github.com/elev1e1nSure/broominal.git
cd broominal

go build -o broominal.exe ./cmd/broominal
.\broominal.exe ui
```

---

## архитектура

```
cmd/broominal/   точка входа cli (cobra)

pkg/
  scanner/       поиск файлов по категориям очистки
  cleaner/       перемещение в карантин и сохранение отчёта
  quarantine/    перемещение, восстановление, очистка, json-манифесты
  report/        генерация json-отчётов
  risk/          классификация риска по путям, расширениям и конфигу
  config/        json-конфигурация и значения по умолчанию
  doctor/        проверки состояния окружения
  i18n/          русская/английская локализация
  style/         ansi-цвета для cli-вывода
  util/          форматирование размеров и общие helpers
  types/         общие доменные типы

internal/
  tui/           интерактивный интерфейс на bubbletea
```

---

## философия

broominal специально скучный. он не обещает чудесную оптимизацию, магию реестра или «ускорение в один клик». он находит кандидатов на очистку, классифицирует риск, показывает результат и перемещает выбранные файлы в карантин, чтобы операцию можно было отменить.

маленькие пакеты. явные зоны ответственности. никакой скрытой магии очистки.

---

## разработка

> включи общие githooks перед коммитами:
> ```powershell
> git config core.hooksPath githooks
> ```

**хуки**
- `pre-commit` — предупреждает, если изменения кода могут требовать обновления документации
- `commit-msg` — проверяет conventional commits

**ci на каждый push / pr в `main`**
```
gofmt → go vet → golangci-lint → go test ./... → сборка windows-артефакта
```

**релиз**
```
git-cliff → сборка broominal.exe → подписанный тег → github release + checksums
```

---

## участие

баг-репорты, идеи новых категорий очистки, улучшения безопасности и странные windows edge cases приветствуются.

см. [CONTRIBUTING.md](CONTRIBUTING.md).

---

## лицензия

[mit](LICENSE) © elev1e1nSure
