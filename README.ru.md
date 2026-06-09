<div align="center">

<h1>Broominal</h1>

<p><strong>Безопасная, прозрачная, отменяемая очистка Windows</strong></p>

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?logo=windows)](https://github.com/elev1e1nSure/broominal)

[🇺🇸 English](README.md)

</div>

---

## Описание

Broominal — это **утилита очистки Windows**, которая никогда не удаляет файлы навсегда. Вместо уничтожения данных, она перемещает их в локальный **карантин** с JSON-манифестом — чтобы вы могли **восстановить всё** в любой момент.

- **Безопасно**: карантин, а не удаление
- **Прозрачно**: отчёты и манифесты в простом JSON
- **Отменяемо**: одна команда для восстановления последней очистки
- **Интерактивно**: удобный TUI с выбором категорий и просмотром файлов
- **Мультиязычно**: английский и русский с автоопределением

---

## Возможности

| Возможность | Описание |
|------------|----------|
| 🧹 **Умное сканирование** | Temp, Downloads, кэш браузеров, Корзина, логи, старые установщики, большие старые файлы, миниатюры (Thumbnails), DirectX Shader Cache, Delivery Optimization, WER, кэш Discord, кэш Steam, кэш VSCode, кэш Edge Code, кэш Chrome Code, Firefox Cache2, старые Temp, .tmp/.log/.bak, пустые папки, npm, pip, Windows Update, дампы, Nvidia, Telegram |
| 🛡️ **Уровни риска** | `safe` / `review` / `danger` — системные пути и `.sys`/`.dll` никогда не трогаются |
| 🔄 **Отмена** | У каждой очистки есть ID восстановления; `restore <id>` возвращает файлы |
| ⚡ **Dry-Run** | `--dry-run` в CLI и клавиша `T` в TUI — посмотреть, что освободится, не трогая файлы |
| ⚙️ **Настройки** | JSON-конфиг: пороги, исключения, вкл/выкл категорий, переопределения риска, язык |
| 🩺 **Doctor** | Встроенные проверки здоровья: права, директории, манифесты, статистика |
| 🗑️ **Очистка карантина** | Автоудаление старых карантинов с предпросмотром `--dry-run` |
| 🖥️ **TUI** | Интерактивный интерфейс на Bubbletea с главным меню, просмотром категорий, восстановлением, доктором, конфигом, выбором языка |
| 🌐 **i18n** | Английский / Русский. Автоопределение языка по IP при первом запуске |

---

## Быстрый старт

### Установка

```powershell
# Из исходников (требуется Go 1.26+)
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest
```

Или скачайте `.exe` из раздела [Releases](../../releases).

### Использование

```powershell
# Сканировать безопасные зоны
broominal scan

# Запустить интерактивный TUI (Главное меню → Сканирование / Восстановление / Доктор / Конфиг / Очистка / Язык)
broominal ui

# Очистить только безопасные элементы
broominal clean --safe

# Симуляция очистки
broominal clean --dry-run

# Восстановить конкретную очистку
broominal restore <id>

# Восстановить с перезаписью существующих файлов
broominal restore <id> --force-overwrite

# Проверки здоровья
broominal doctor

# Показать конфиг
broominal config

# Удалить карантины старше 30 дней
broominal quarantine-cleanup --dry-run
broominal quarantine-cleanup --force
broominal quarantine-cleanup --force --max-age-days 7
```

---

## Архитектура

```
cmd/broominal/      Точка входа CLI (Cobra)
pkg/
  scanner/          Поиск файлов по категориям (25+ целей сканирования)
  quarantine/       Перемещение / Восстановление / Очистка карантина с манифестами JSON
  report/           Генерация JSON-отчётов
  risk/             Классификация риска (путь, расширение, конфиг)
  config/           JSON-конфигурация (пороги, исключения, переопределения, язык)
  doctor/           Проверки здоровья (права, директории, манифесты, статистика)
  i18n/             Локализация (EN/RU, автоопределение, T-ключи)
  style/            ANSI-цвета для CLI-вывода
  types/            Общие доменные типы
internal/
  tui/              Интерактивный интерфейс на Bubbletea (Главное меню → несколько экранов)
```

## Лицензия

[MIT](LICENSE) © elev1e1nSure
