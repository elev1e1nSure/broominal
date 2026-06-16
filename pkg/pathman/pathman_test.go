package pathman

import (
	"reflect"
	"testing"
)

func TestPathEntries(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "empty",
			path:     "",
			expected: nil,
		},
		{
			name:     "single",
			path:     "C:\\Windows",
			expected: []string{"C:\\Windows"},
		},
		{
			name:     "multiple with empty segments",
			path:     "C:\\Windows;;D:\\Apps;  ;E:\\Tools;",
			expected: []string{"C:\\Windows", "D:\\Apps", "E:\\Tools"},
		},
		{
			name:     "trim spaces",
			path:     "  C:\\Windows  ; D:\\Apps ",
			expected: []string{"C:\\Windows", "D:\\Apps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathEntries(tt.path)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("pathEntries() = %v, want %v", got, tt.expected)
			}
		})
	}
}
