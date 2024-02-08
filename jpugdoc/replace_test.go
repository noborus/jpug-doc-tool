package jpugdoc

import (
	"reflect"
	"testing"

	"github.com/noborus/go-textra"
)

func TestRep_replaceCatalogs(t *testing.T) {
	type fields struct {
		catalogs []Catalog
		vTag     string
		update   bool
		mt       bool
		prompt   bool
		similar  int
		api      *textra.TexTra
		apiType  string
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
			name: "Test replaceCatalogs",
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
				mt:       tt.fields.mt,
				prompt:   tt.fields.prompt,
				similar:  tt.fields.similar,
				api:      tt.fields.api,
				apiType:  tt.fields.apiType,
			}
			if got := rep.replaceCatalogs(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rep.replaceCatalogs() = \n%v\nwant \n%v\n", string(got), string(tt.want))
			}
		})
	}
}
