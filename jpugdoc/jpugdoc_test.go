package jpugdoc

import (
	"reflect"
	"testing"
)

func Test_version(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test16.0",
			args: args{
				[]byte(`<!ENTITY version "16.0">`),
			},
			want:    "REL_16_0",
			wantErr: false,
		},
		{
			name: "test17devel",
			args: args{
				[]byte(`<!ENTITY version "17devel">`),
			},
			want:    "master",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := version(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("version() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCatalogs(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want []Catalog
	}{
		{
			name: "testTitle",
			args: args{
				src: []byte(`␝␟ <title>Acronyms</title>␟ <title>頭字語</title>␞␞␞`),
			},
			want: []Catalog{
				{
					pre:      "",
					en:       " <title>Acronyms</title>",
					ja:       " <title>頭字語</title>",
					preCDATA: "",
					post:     "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCatalogs(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCatalogs() = %v, want %v", got, tt.want)
			}
		})
	}
}
