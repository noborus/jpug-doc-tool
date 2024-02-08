package jpugdoc

import (
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
