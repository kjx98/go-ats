package ats

import (
	"testing"
)

func TestTimeT32(t *testing.T) {
	tests := []struct {
		name  string
		timeV timeT32
		want  int64
	}{
		// TODO: Add test cases.
		{"timeT32test1", 1, 1},
		{"timeT32test2", 0xffffffff, (1 << 32) - 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.timeV.Unix(); got != tt.want {
				t.Errorf("timeT32.Unix() = %v, want %v", got, tt.want)
			}
		})
	}
	var tt timeT32
	t.Log("dooms time", tt.Dooms().Format("2006-01-02 15:04:05"))
}
