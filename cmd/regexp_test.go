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

func Test_splitComment(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name   string
		args   args
		wantEn []byte
		wantJa []byte
		wantEx []byte
	}{
		{
			name: "test1",
			args: args{
				[]byte(`<para>
<!--
  comment
-->
日本語
</para>`),
			},
			wantEn: []byte("\n  comment\n"),
			wantJa: []byte("\n日本語\n"),
			wantEx: []byte("</para>"),
		},
		{
			name: "test2",
			args: args{
				[]byte(`<para>
<!--
  comment
-->
日本語
<orderedlist spacing="compact">`),
			},
			wantEn: []byte("\n  comment\n"),
			wantJa: []byte("\n日本語\n"),
			wantEx: []byte(`<orderedlist`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEn, gotJa, gotEx := splitComment(tt.args.src)
			if !reflect.DeepEqual(gotEn, tt.wantEn) {
				t.Errorf("splitComment() gotEn = %v, want %v", gotEn, tt.wantEn)
			}
			if !reflect.DeepEqual(gotJa, tt.wantJa) {
				t.Errorf("splitComment() gotJa = %v, want %v", gotJa, tt.wantJa)
			}
			if !reflect.DeepEqual(gotEx, tt.wantEx) {
				t.Errorf("splitComment() gotEx = %s, want %s", gotEx, tt.wantEx)
			}
		})
	}
}
