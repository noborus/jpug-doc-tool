package jpugdoc

import "testing"

func TestPostProcess(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "test1",
			input: "テスト。てすと。",
			want:  "テスト。\nてすと。",
		},
		{
			name:  "test2",
			input: "テスト。（てすと）。",
			want:  "テスト。\n（てすと）。",
		},
		{
			name:  "test3",
			input: "テスト。（てすと。）",
			want:  "テスト。\n（てすと。）",
		},
		{
			name:  "test4",
			input: "テスト。(test)。",
			want:  "テスト。\n(test)。",
		},
		{
			name:  "test5",
			input: "テスト。(テスト)。",
			want:  "テスト。\n（テスト）。",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := postProcess(tt.input)
			if got != tt.want {
				t.Errorf("postProcess() = [%s], want [%s]", got, tt.want)
			}
		})
	}
}
