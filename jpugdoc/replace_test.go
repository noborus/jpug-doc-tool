package jpugdoc

import (
	"reflect"
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
		src   []byte
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
				src:   []byte("This is a test1.\n"),
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
				src:   []byte("This is a test.\n"),
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
			got, got1 := rep.findSimilar(tt.args.src, tt.args.enStr)
			if got != tt.want {
				t.Errorf("Rep.findSimilar() got = %v, want %v", got, tt.want)
			}
			if got1 < tt.want1 {
				t.Errorf("Rep.findSimilar() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
