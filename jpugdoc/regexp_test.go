package jpugdoc

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
				[]byte("    <para>\ntest\n</para>"),
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
				[]byte("<!--\n<para>test</para>\n-->"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParaAll(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParaAll() = %s, want %s", got, tt.want)
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

func Test_stripPROGRAMLISTING(t *testing.T) {
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
				[]byte(`a<programlisting>16</programlisting>b`),
			},
			want: []byte("ab"),
		},
		{
			name: "test2",
			args: args{
				[]byte(`a<screen>16</screen>b`),
			},
			want: []byte("ab"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripPROGRAMLISTING(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripPROGRAMLISTING() = [%v], want [%v]", string(got), string(tt.want))
			}
		})
	}
}

func Test_regParaScreen(t *testing.T) {
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
				src: []byte(`<para>
<screen>
test
</screen>
</para>`),
			},
			want: [][]byte{
				[]byte(`<para>
<screen>`),
			},
		},
		{
			name: "test2",
			args: args{
				src: []byte(`  <para>
  A custom scan provider will typically add paths for a base relation by
<programlisting>
  typedef void (*set_rel_pathlist_hook_type) (PlannerInfo *root,
									RelOptInfo *rel,
										Index rti,
									RangeTblEntry *rte);
  extern PGDLLIMPORT set_rel_pathlist_hook_type set_rel_pathlist_hook;
</programlisting>
  </para>`),
			},
			want: [][]byte{
				[]byte(`<para>
  A custom scan provider will typically add paths for a base relation by
<programlisting>`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := regParaScreen(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				if len(got) != len(tt.want) {
					t.Errorf("regParaScreen(): len(got) = %d, len(want) = %d", len(got), len(tt.want))
				}
				for i := range got {
					t.Errorf("regParaScreen():%d = %v, want %v", i, string(got[i]), string(tt.want[i]))
				}
			}
		})
	}
}

func Test_similarBlank(t *testing.T) {
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
				src: []byte(`<!--
te,st
-->
《マッチ度[]》
`),
			},
			want: [][]byte{
				[]byte(`<!--
te,st
-->
《マッチ度[]》
`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := similarBlank(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Error("got:", got)
				if len(got) != len(tt.want) {
					t.Errorf("similarBlank(): len(got) = %d, len(want) = %d", len(got), len(tt.want))
				}
				for i := range got {
					t.Errorf("similarBlank():%d = %v, want %v", i, string(got[i]), string(tt.want[i]))
				}
			}
		})
	}
}

func Test_signMaatch(t *testing.T) {
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
				src: []byte("test(en)"),
			},
			want: []byte("(en)"),
		},
		{
			name: "Heikki Linnakangas",
			args: args{
				src: []byte("Fix (Heikki Linnakangas)"),
			},
			want: []byte("(Heikki Linnakangas)"),
		},
		{
			name: "test2",
			args: args{
				src: []byte(`(Andres Freund,
				Daniel Gustafsson)`),
			},
			want: []byte("(Andres Freund, Daniel Gustafsson)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := signMatch(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("signMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
