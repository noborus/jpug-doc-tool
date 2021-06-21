package cmd

import (
	"reflect"
	"testing"
)

func Test_paraAll(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "test1",
			args: args{
				src: []byte("    <para>\ntest\n</para>   "),
			},
			want: [][]byte{
				[]byte("<para>\ntest\n</para>"),
			},
		},
		{
			name: "test2",
			args: args{
				src: []byte("<para>\n<!--\ncomment\n-->\ntest\n</para>"),
			},
			want: [][]byte{
				[]byte("<para>\n<!--\ncomment\n-->\ntest\n</para>"),
			},
		},
		{
			name: "test3",
			args: args{
				src: []byte("<para>test</para>"),
			},
			want: [][]byte{
				[]byte("<para>test</para>"),
			},
		},
		{
			name: "test4",
			args: args{
				src: []byte(`<!--
<para>test</para>
-->`),
			},
			want: [][]byte{
				[]byte("<!--\n<para>test</para>"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := paraAll(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("paraAll() = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_stripNONJA(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test1",
			args: args{
				src: []byte(`
あいうえお
`),
			},
			want: []byte(`
あいうえお`),
		},
		{
			name: "test2",
			args: args{
				src: []byte(`
あいうえお
<literal>a</literal>
`),
			},
			want: []byte(`
あいうえお`),
		},
		{
			name: "test3",
			args: args{
				src: []byte(`OpenSSLはよく使われる曲線に名前を付けています。
		<literal>prime256v1</literal> (NIST P-256),
        <literal>secp384r1</literal> (NIST P-384),
        <literal>secp521r1</literal> (NIST P-521).`),
			},
			want: []byte(`OpenSSLはよく使われる曲線に名前を付けています。`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripNONJA(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripNONJA() = %s, want %s", got, tt.want)
			}
		})
	}
}
