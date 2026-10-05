package jpugdoc

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestRep_matchReplace(t *testing.T) {
	type fields struct {
		catalogs []Catalog
		vTag     string
		update   bool
		prompt   bool
		similar  int
	}
	type args struct {
		src []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "Test matchReplace1",
			fields: fields{
				catalogs: []Catalog{
					{
						en: "This is a test.",
						ja: "これはテストです。",
					},
				},
			},
			args: args{
				src: []byte("This is a test.\n"),
			},
			want: []byte(`<!--
This is a test.
-->
これはテストです。
`),
		},
		{
			name: "no replace",
			fields: fields{
				catalogs: []Catalog{
					{
						en: "This is a test.",
						ja: "これはテストです。",
					},
				},
			},
			args: args{
				src: []byte("That is a test.\n"),
			},
			want: []byte("That is a test.\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := Rep{
				catalogs: tt.fields.catalogs,
				vTag:     tt.fields.vTag,
				update:   tt.fields.update,
				prompt:   tt.fields.prompt,
				similar:  tt.fields.similar,
			}
			if got := rep.matchReplace(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rep.matchReplace() = \n%v\nwant \n%v\n", string(got), string(tt.want))
			}
		})
	}
}

func TestRep_matchReplaceCommonSkipsCommentedText(t *testing.T) {
	rep := Rep{
		catalogs: []Catalog{
			{
				en: "First line.\nCommon phrase.",
				ja: "翻訳文",
			},
		},
		common: regCompile(Catalogs{
			{
				en: "Common phrase.",
				ja: "共通訳",
			},
		}),
	}
	src := []byte("First line.\nCommon phrase.\n")
	want := []byte("<!--\nFirst line.\nCommon phrase.\n-->\n翻訳文\n")

	if got := rep.matchReplace(src); !reflect.DeepEqual(got, want) {
		t.Errorf("Rep.matchReplace() = \n%s\nwant \n%s\n", string(got), string(want))
	}
}

func TestBlockReplaceUlinkPreservesLinkLines(t *testing.T) {
	rep := Rep{
		similar: 1,
		mt:      100,
		catalogs: []Catalog{
			{
				en: "Reject calls from SQL to functions that take or return type <type>internal</type> (Tom Lane)",
				ja: "SQLから<type>internal</type>型を取る、または返す関数の呼び出しを拒否します。(Tom Lane)",
			},
		},
	}
	src := "<para>\n" +
		"  Reject calls from SQL to functions that take or return\n" +
		"  type <type>internal</type> (Tom Lane)\n" +
		"  \n" +
		"  <ulink url=\"&commit_baseurl;eb9e55297\">&sect;</ulink>\n" +
		"  <ulink url=\"&commit_baseurl;83d0a083f\">&sect;</ulink>\n" +
		" </para>"

	got := rep.blockReplace(src)
	commentEnd := strings.Index(got, "\n-->")
	if commentEnd < 0 {
		t.Fatalf("blockReplace() output has no comment end:\n%s", got)
	}
	if got == src {
		t.Fatal("blockReplace() did not replace the matching paragraph")
	}
	if strings.HasSuffix(got[:commentEnd], "\n  ") {
		t.Errorf("comment contains a trailing whitespace-only line:\n%s", got)
	}
	linkLines := "  <ulink url=\"&commit_baseurl;eb9e55297\">&sect;</ulink>\n" +
		"  <ulink url=\"&commit_baseurl;83d0a083f\">&sect;</ulink>\n </para>"
	if n := strings.Count(got, "<ulink"); n != 2 {
		t.Errorf("ulink count = %d, want 2:\n%s", n, got)
	}
	if n := strings.Count(got, "</para>"); n != 1 {
		t.Errorf("</para> count = %d, want 1:\n%s", n, got)
	}
	if !strings.HasSuffix(got, linkLines) {
		t.Errorf("ulink lines changed or moved:\n%s", got)
	}
}

func TestBlockReplaceCVELineExcluded(t *testing.T) {
	rep := Rep{
		similar: 1,
		mt:      100,
		catalogs: []Catalog{
			{
				en: "Fix a bug in the parser.",
				ja: "パーサのバグを修正します。",
			},
		},
	}
	src := "<para>\n" +
		"      Fix a bug in the parser.\n" +
		"      (CVE-2026-6472)\n" +
		"     </para>"

	got := rep.blockReplace(src)
	if got == src {
		t.Fatal("blockReplace() did not replace the paragraph")
	}
	if !strings.HasSuffix(got, "\n      (CVE-2026-6472)\n     </para>") {
		t.Errorf("CVE line changed or moved:\n%s", got)
	}
	if n := strings.Count(got, "CVE-2026-6472"); n != 1 {
		t.Errorf("CVE count = %d, want 1:\n%s", n, got)
	}
	if end := strings.Index(got, "-->"); strings.Contains(got[:end], "CVE-") {
		t.Errorf("CVE line is inside the comment:\n%s", got)
	}
}

func TestRep_findSimilar(t *testing.T) {
	type fields struct {
		catalogs []Catalog
		vTag     string
		update   bool
		similar  int
		mt       int
		prompt   bool
	}
	type args struct {
		enStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  float64
	}{
		{
			name: "Test findSimilar1",
			fields: fields{
				catalogs: []Catalog{
					{
						en: "This is a test.",
						ja: "これはテストです。",
					},
				},
				similar: 50,
			},
			args: args{
				enStr: "This is a test1.",
			},
			want:  "これはテストです。",
			want1: 50,
		},
		{
			name: "Test findSimilarMatch",
			fields: fields{
				catalogs: []Catalog{
					{
						en: "This is a test.",
						ja: "《マッチ度 99》これはテストです。",
					},
				},
				similar: 50,
			},
			args: args{
				enStr: "This is a test.",
			},
			want:  "これはテストです。",
			want1: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := &Rep{
				catalogs: tt.fields.catalogs,
				vTag:     tt.fields.vTag,
				update:   tt.fields.update,
				similar:  tt.fields.similar,
				mt:       tt.fields.mt,
				prompt:   tt.fields.prompt,
			}
			got, got1 := rep.findSimilar(tt.args.enStr)
			if got != tt.want {
				t.Errorf("Rep.findSimilar() got = %v, want %v", got, tt.want)
			}
			if got1 < tt.want1 {
				t.Errorf("Rep.findSimilar() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMatchCommon(t *testing.T) {
	tests := []struct {
		name    string
		src     []byte
		catalog Catalog
		want    []byte
	}{
		{
			name: "Match and replace text",
			src:  []byte("   This is a test.\nAnother line."),
			catalog: Catalog{
				en:        "This is a test.",
				ja:        " 新しいテキスト",
				commonReg: regexp.MustCompile(`(?s)(\s*)This is a test.\n`),
			},
			want: []byte("   <!--\nThis is a test.\n-->\n 新しいテキスト\nAnother line."),
		},
		{
			name: "No match, no replacement",
			src:  []byte("No matching text here.\nAnother line."),
			catalog: Catalog{
				en:        "This is a test.",
				ja:        "新しいテキスト",
				commonReg: regexp.MustCompile(`(?s)(\s*)This is a test.\n`),
			},
			want: []byte("No matching text here.\nAnother line."),
		},
		{
			name: "Multiple matches",
			src:  []byte("This is a test.\nThis is a test.\nAnother line."),
			catalog: Catalog{
				en:        "This is a test.",
				ja:        "新しいテキスト",
				commonReg: regexp.MustCompile(`(?s)(\s*)This is a test.\n`),
			},
			want: []byte("<!--\nThis is a test.\n-->\n新しいテキスト\n<!--\nThis is a test.\n-->\n新しいテキスト\nAnother line."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchCommon(tt.src, tt.catalog)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchCommon() = \n%s, want \n%s", string(got), string(tt.want))
			}
		})
	}
}
