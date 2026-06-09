package style

import "fmt"

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

func Sprintf(color, format string, a ...any) string {
	return fmt.Sprintf(color+format+Reset, a...)
}

func Boldf(format string, a ...any) string {
	return Sprintf(Bold, format, a...)
}

func Greenf(format string, a ...any) string {
	return Sprintf(Green, format, a...)
}

func Yellowf(format string, a ...any) string {
	return Sprintf(Yellow, format, a...)
}

func Redf(format string, a ...any) string {
	return Sprintf(Red, format, a...)
}

func Cyanf(format string, a ...any) string {
	return Sprintf(Cyan, format, a...)
}

func Grayf(format string, a ...any) string {
	return Sprintf(Gray, format, a...)
}

func Passf(format string, a ...any) string {
	return Sprintf(Green+Bold, format, a...)
}

func Warnf(format string, a ...any) string {
	return Sprintf(Yellow+Bold, format, a...)
}

func Failf(format string, a ...any) string {
	return Sprintf(Red+Bold, format, a...)
}
