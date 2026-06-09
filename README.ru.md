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

---

## Возможности

| Возможность | Описание |
|------------|----------|
| 🧹 **Умное сканирование** | Temp, Downloads, кэш браузеров, Корзина, логи, старые установщики, большие старые файлы |
| 🛡️ **Уровни риска** | `safe` / `review` / `danger` — системные пути и `.sys`/`.dll` никогда не трогаются |
| 🔄 **Отмена** | У каждой очистки есть ID восстановления; `restore last` возвращает файлы |
| ⚡ **Dry-Run** | `--dry-run` в CLI и клавиша `T` в TUI — посмотреть, что освободится, не трогая файлы |
| ⚙️ **Настройки** | JSON-конфиг: пороги, исключения, вкл/выкл категорий, переопределения риска |
| 🩺 **Doctor** | Встроенные проверки здоровья: права, манифесты, место на диске |
| 🗑️ **Очистка карантина** | Автоудаление старых карантинов с предпросмотром `--dry-run` |
| 🖥️ **TUI** | Интерактивный интерфейс на Bubbletea с просмотром категорий и файлов |

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

# Запустить интерактивный TUI
broominal ui

# Очистить только безопасные элементы
broominal clean --safe

# Симуляция очистки
broominal clean --dry-run

# Восстановить последнюю очистку
broominal restore last

# Восстановить с перезаписью существующих файлов
broominal restore last --force-overwrite

# Проверки здоровья
broominal doctor

# Показать конфиг
broominal config

# Удалить карантины старше 30 дней
broominal quarantine-cleanup --dry-run
broominal quarantine-cleanup --force
```

---

## Архитектура

```
cmd/broominal/      Точка входа CLI (Cobra)
pkg/
  scanner/          Поиск файлов по категориям
  quarantine/       Перемещение / Восстановление / Очистка карантина с манифестами JSON
  report/           Генерация JSON-отчётов
  risk/             Классификация риска (путь, расширение, конфиг)
  config/           JSON-конфигурация (пороги, исключения, переопределения)
  types/            Общие доменные типы
internal/
  tui/              Интерактивный интерфейс на Bubbletea
```

## Лицензия

[MIT](LICENSE) © elev1e1nSure
