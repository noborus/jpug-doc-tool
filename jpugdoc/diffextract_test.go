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
			name: "test1",
			args: args{
				[]byte(` 
 
 
 
 <para>
+<!--
 test
+-->
+テスト
 </para>`),
			},
			want: []Catalog{
				{
					en: "test",
					ja: "テスト",
				},
			},
		},
		{
			name: "test2",
			args: args{
				[]byte(` 
 
 
 
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
					pre:      "",
					en:       " test",
					ja:       "テスト",
					preCDATA: "",
				},
				{
					pre:      "",
					en:       " test2",
					ja:       "テスト２",
					preCDATA: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Extraction(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Extraction() = %#v, want %#v", got, tt.want)
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
					pre:      "",
					en:       "   <title>Accessing a Database</title>",
					ja:       "   <title>データベースへのアクセス</title>",
					preCDATA: "",
				},
				/*
									{
										pre: `   <indexterm zone="tutorial-accessdb">
					    <primary>psql</primary>
					   </indexterm>
					`,
										ja: "    <indexterm><primary>スーパーユーザ</primary></indexterm>",
									},
				*/
				{
					pre:      "<!--",
					en:       "    <indexterm><primary>superuser</primary></indexterm>\n    The last line could also be:",
					ja:       "    <indexterm><primary>スーパーユーザ</primary></indexterm>\n最後の行は以下のようになっているかもしれません。",
					preCDATA: "",
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
