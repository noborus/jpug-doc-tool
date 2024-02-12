package jpugdoc

import (
	"io"
	"os"
	"reflect"
	"testing"
)

func TestExtraction(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want []Catalog
	}{
		{
			name: "testNormal",
			args: args{
				[]byte(` 
@@ 
 <para>
+<!--
 test
+-->
+テスト
 </para>`),
			},
			want: []Catalog{
				{
					pre:      "<para>",
					en:       "test",
					ja:       "テスト",
					preCDATA: "",
					post:     "</para>",
				},
			},
		},
		{
			name: "testDouble",
			args: args{
				[]byte(` 
@@
 <para>
+<!--
  test
+-->
+テスト
+<!--
  test2
+-->
+テスト２
 </para>`),
			},
			want: []Catalog{
				{
					pre:      "<para>",
					en:       " test",
					ja:       "テスト",
					preCDATA: "",
					post:     "",
				},
				{
					pre:      " test",
					en:       " test2",
					ja:       "テスト２",
					preCDATA: "",
					post:     "</para>",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Extraction(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Extraction() = \n%#v, want \n%#v", got, tt.want)
			}
		})
	}
}

func TestExtraction2(t *testing.T) {
	type args struct {
		diffSrc []byte
	}
	tests := []struct {
		name string
		args args
		want []Catalog
	}{
		{
			name: "Test with indexterm.diff",
			args: func() args {
				file, err := os.Open("../testdata/indexterm.diff")
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()

				bytes, err := io.ReadAll(file)
				if err != nil {
					t.Fatal(err)
				}
				return args{diffSrc: bytes}
			}(),
			want: []Catalog{
				{
					pre:  `  <sect1 id="tutorial-accessdb">`,
					en:   "   <title>Accessing a Database</title>",
					ja:   "   <title>データベースへのアクセス</title>",
					post: "",
				},
				{
					pre:  "",
					en:   "    <indexterm><primary>superuser</primary></indexterm>\n    The last line could also be:",
					ja:   "    <indexterm><primary>スーパーユーザ</primary></indexterm>\n最後の行は以下のようになっているかもしれません。",
					post: "",
				},
			},
		},
		{
			name: "Test width row.diff",
			args: func() args {
				file, err := os.Open("../testdata/row.diff")
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()

				bytes, err := io.ReadAll(file)
				if err != nil {
					t.Fatal(err)
				}
				return args{diffSrc: bytes}
			}(),
			want: []Catalog{
				{
					pre:  "     <row>",
					en:   "      <entry>External Syntax</entry>",
					ja:   "      <entry>外部構文</entry>",
					post: "      <entry>Meaning</entry>",
				},
				{
					pre:  "      <entry>External Syntax</entry>",
					en:   "      <entry>Meaning</entry>",
					ja:   "      <entry>意味</entry>",
					post: `     </row>`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Extraction(tt.args.diffSrc); !reflect.DeepEqual(got, tt.want) {
				t.Logf("len: %v!=%v\n", len(got), len(tt.want))
				for i, v := range got {
					if !reflect.DeepEqual(v.pre, tt.want[i].pre) {
						t.Errorf("Extraction().pre = [%s]\n, want [%s]\n", v.pre, tt.want[i].pre)
					}
					if !reflect.DeepEqual(v.en, tt.want[i].en) {
						t.Errorf("Extraction().en = %v\n, want %v\n", v.en, tt.want[i].en)
					}
					if !reflect.DeepEqual(v.ja, tt.want[i].ja) {
						t.Errorf("Extraction().ja = %v\n, want %v\n", v.ja, tt.want[i].ja)
					}
					if !reflect.DeepEqual(v.preCDATA, tt.want[i].preCDATA) {
						t.Errorf("Extraction().CDATA = %v\n, want %v\n", v.preCDATA, tt.want[i].preCDATA)
					}
				}
				//t.Errorf("Extraction() = %s\n, want %s\n", got, tt.want)
			}
		})
	}
}

func Test_trimPrefix(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				s: []string{
					"a",
					"",
					"",
					"",
					"<para>",
					"<!--",
					" test",
					"-->",
					"テスト",
					"</para>",
				},
			},
			want: []string{
				"<para>",
				"<!--",
				" test",
				"-->",
				"テスト",
				"</para>",
			},
		},
		{
			name: "test2",
			args: args{
				s: []string{
					"<para>",
					"<!--",
					" test",
					"-->",
					"テスト",
					"</para>",
				},
			},
			want: []string{
				"<para>",
				"<!--",
				" test",
				"-->",
				"テスト",
				"</para>",
			},
		},
		{
			name: "test3",
			args: args{
				s: []string{
					"header",
					"",
					"<para>",
					"<!--",
					" test",
					"-->",
					"テスト",
					"</para>",
				},
			},
			want: []string{
				"<para>",
				"<!--",
				" test",
				"-->",
				"テスト",
				"</para>",
			},
		},
		{
			name: "test4",
			args: args{
				s: []string{
					"hello",
					"world",
					"",
				},
			},
			want: []string{
				"hello",
				"world",
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prefixBlock(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trimPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
