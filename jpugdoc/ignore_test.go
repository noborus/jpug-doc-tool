package jpugdoc

import (
	"bytes"
	"strings"
	"testing"
)

func Test_writeIgnore(t *testing.T) {
	type args struct {
		ignores []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Test_writeIgnore",
			args:    args{[]string{"a", "b"}},
			want:    "a\nb\n",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := new(bytes.Buffer)
			if err := writeIgnore(buffer, tt.args.ignores); (err != nil) != tt.wantErr {
				t.Errorf("writeIgnore() error = %v, wantErr %v", err, tt.wantErr)
			}
			if buffer.String() != tt.want {
				t.Errorf("writeIgnore() = %v, want %v", buffer.String(), tt.want)
			}
		})
	}
}

func Test_readIgnore(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IgnoreList
	}{
		{
			name:     "Test_readIgnore",
			input:    "a\nb\nc\n",
			expected: IgnoreList{"a": true, "b": true, "c": true},
		},
		{
			name:     "Test_readIgnore_Empty",
			input:    "",
			expected: IgnoreList{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			ignoreList := readIgnore(reader)

			if len(ignoreList) != len(tt.expected) {
				t.Errorf("readIgnore() returned ignore list of length %d, expected %d", len(ignoreList), len(tt.expected))
			}

			for key := range tt.expected {
				if !ignoreList[key] {
					t.Errorf("readIgnore() did not contain expected key: %s", key)
				}
			}
		})
	}
}
