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

func Test_enjaCheck(t *testing.T) {
	type args struct {
		fileName string
		catalog  Catalog
		cf       CheckFlag
	}
	tests := []struct {
		name    string
		args    args
		want    []result
		wantErr bool
	}{
		{
			name: "testWordCheck",
			args: args{
				fileName: "test",
				catalog: Catalog{
					pre: "",
					en:  "test",
					ja:  "tetテスト",
				},
				cf: CheckFlag{
					VTag:   "REL_16_0",
					Ignore: false,
					WIP:    false,
					Para:   true,
					Word:   true,
					Tag:    true,
					Num:    true,
					Strict: false,
				},
			},
			want: []result{
				{
					comment: "[tet]が含まれていません",
					en:      "test",
					ja:      "tetテスト",
				},
			},
			wantErr: false,
		},
		{
			name: "testTagCheck",
			args: args{
				fileName: "test",
				catalog: Catalog{
					pre: "",
					en:  "<literal>test</literal>",
					ja:  "<quote>テスト</quote>",
				},
				cf: CheckFlag{
					VTag:   "REL_16_0",
					Ignore: false,
					WIP:    false,
					Para:   true,
					Word:   true,
					Tag:    true,
					Num:    true,
					Strict: false,
				},
			},
			want: []result{
				{
					comment: "[quote]が含まれていません",
					en:      "<literal>test</literal>",
					ja:      "<quote>テスト</quote>",
				},
				{
					comment: "タグ[(<literal>)1:0 ｜ (</literal>)1:0]の数が違います",
					en:      "<literal>test</literal>",
					ja:      "<quote>テスト</quote>",
				},
			},
			wantErr: false,
		},
		{
			name: "testNumCheck",
			args: args{
				fileName: "test",
				catalog: Catalog{
					pre: "",
					en:  "PostgreSQL 9.6",
					ja:  "PostgreSQL 10.1",
				},
				cf: CheckFlag{
					VTag:   "REL_16_0",
					Ignore: false,
					WIP:    false,
					Para:   true,
					Word:   true,
					Tag:    true,
					Num:    true,
					Strict: false,
				},
			},
			want: []result{
				{
					comment: "原文にある[9 ｜ 6]が含まれていません",
					en:      "PostgreSQL 9.6",
					ja:      "PostgreSQL 10.1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enjaCheck(tt.args.fileName, tt.args.catalog, tt.args.cf)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("enjaCheck() = \n%v\n, want\n%v\n", got, tt.want)
			}
		})
	}
}
