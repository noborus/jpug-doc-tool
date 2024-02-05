package jpugdoc

import (
	"reflect"
	"testing"
)

func Test_numOfTagCheck(t *testing.T) {
	type args struct {
		strict bool
		en     string
		ja     string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test",
			args: args{
				strict: true,
				en:     "<productname>PostgreSQL</productname> startup script",
				ja:     "<productname>PostgreSQL</productname>スタートアップスクリプト",
			},
			want: []string{},
		},
		{
			name: "testFail",
			args: args{
				strict: false,
				en:     "<productname>PostgreSQL</productname> startup script",
				ja:     "PostgreSQLスタートアップスクリプト",
			},
			want: []string{
				"(<productname>)1:0",
				"(</productname>)1:0",
			},
		},
		{
			name: "testFailStrict",
			args: args{
				strict: true,
				en:     "<productname>PostgreSQL</productname> startup script",
				ja:     "<productname>PostgreSQL</productname>スタートアップスクリプト<productname>PostgreSQL</productname>スタートアップスクリプト",
			},
			want: []string{
				"(<productname>)1:2",
				"(</productname>)1:2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := numOfTagCheck(tt.args.strict, tt.args.en, tt.args.ja); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("numOfTagCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_numCheck(t *testing.T) {
	type args struct {
		en string
		ja string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test",
			args: args{
				en: "PostgreSQL 9.6",
				ja: "PostgreSQL 9.6",
			},
			want: []string{},
		},
		{
			name: "testFail",
			args: args{
				en: "PostgreSQL 9.6",
				ja: "PostgreSQL 10.1",
			},
			want: []string{
				"9",
				"6",
			},
		},
		{
			name: "testIgnore",
			args: args{
				en: "--",
				ja: "-- #45",
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := numCheck(tt.args.en, tt.args.ja); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("numCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_wordCheck(t *testing.T) {
	type args struct {
		en string
		ja string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test",
			args: args{
				en: "PostgreSQL",
				ja: "PostgreSQL",
			},
			want: []string{},
		},
		{
			name: "testFail",
			args: args{
				en: "execute command",
				ja: "SQLコマンドを実行する",
			},
			want: []string{
				"SQL",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wordCheck(tt.args.en, tt.args.ja); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wordCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
