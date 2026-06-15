package scanner

import (
	"testing"
)

func FuzzExtractTaskCommand(f *testing.F) {
	f.Add([]byte(`<Task><Exec><Command>notepad.exe</Command></Exec></Task>`))
	f.Add([]byte{0xFF, 0xFE, '<', 0, 'C', 0, 'o', 0, 'm', 0, 'm', 0, 'a', 0, 'n', 0, 'd', 0, '>', 0, 'c', 0, 'm', 0, 'd', 0, '.', 0, 'e', 0, 'x', 0, 'e', 0, '<', 0, '/', 0, 'C', 0, 'o', 0, 'm', 0, 'm', 0, 'a', 0, 'n', 0, 'd', 0, '>', 0})
	f.Add([]byte(``))
	f.Add([]byte(`<Task></Task>`))
	f.Add([]byte(string(make([]byte, 10000))))

	f.Fuzz(func(t *testing.T, data []byte) {
		cmd := extractTaskCommand(data)
		if len(cmd) > 10000 {
			t.Errorf("extractTaskCommand returned excessively long command: %d bytes", len(cmd))
		}
		if len(data) == 0 && cmd != "" {
			t.Errorf("extractTaskCommand returned non-empty for empty input: %q", cmd)
		}
	})
}
