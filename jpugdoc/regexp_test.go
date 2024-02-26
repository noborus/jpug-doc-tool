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
《》
`),
			},
			want: [][]byte{
				[]byte(`<!--
te,st
-->
《》
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

func Test_authorMatch(t *testing.T) {
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
		{
			name: "test3",
			args: args{
				src: []byte(`Return the correct status code when a new client disconnects without
 responding to the server's password challenge (Liu Lang, Tom Lane)
 `),
			},
			want: []byte("(Liu Lang, Tom Lane)"),
		},
		{
			name: "test4",
			args: args{
				src: []byte(`(Franz-Josef Färber)`),
			},
			want: []byte("(Franz-Josef Färber)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := authorMatch(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("authorMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_titleMatch(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test no title",
			args: args{
				src: []byte(`test`),
			},
			want: nil,
		},
		{
			name: "test1",
			args: args{
				src: []byte(`<!-- doc/src/sgml/wal.sgml -->

<chapter id="wal">
 <title>Reliability and the Write-Ahead Log</title>

 <para>
`),
			},
			want: []byte(` <title>Reliability and the Write-Ahead Log</title>`),
		},
		{
			name: "test2",
			args: args{
				src: []byte(`<!-- doc/src/sgml/wal.sgml -->

<chapter id="wal">
<!--
 <title>Reliability and the Write-Ahead Log</title>
-->
 <title>信頼性とログ先行書き込み</title>

 <para>
`),
			},
			want: nil,
		},
		{
			name: "test3",
			args: args{
				src: []byte(`
<sect2 id="bloom-examples">
 <title>Examples</title>
			   
<para>
`),
			},
			want: []byte(` <title>Examples</title>`),
		},
		{
			name: "test4",
			args: args{
				src: []byte(` <sect1 id="release-16-1">
  <title>Release 16.1</title>

  <para>
 `),
			},
			want: []byte(` <title>Release 16.1</title`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := titleMatch(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("titleMatch() = [%v], want [%v]", string(got), string(tt.want))
			}
		})
	}
}

func Test_titleMatch2(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name  string
		args  args
		want1 []byte
		want2 []byte
	}{
		{
			name: "test no title",
			args: args{
				src: []byte(`test`),
			},
			want1: nil,
			want2: nil,
		},
		{
			name: "test1",
			args: args{
				src: []byte(`<!-- doc/src/sgml/wal.sgml -->

<chapter id="wal">
 <title>Reliability and the Write-Ahead Log</title>
 
<para>
`),
			},
			want1: []byte(nil),
			want2: []byte(nil),
		},
		{
			name: "test2",
			args: args{
				src: []byte(`<!-- doc/src/sgml/wal.sgml -->

<chapter id="wal">
<!--
 <title>Reliability and the Write-Ahead Log</title>
-->
 <title>信頼性とログ先行書き込み</title>
 
<para>
`),
			},
			want1: []byte("<title>Reliability and the Write-Ahead Log</title>"),
			want2: []byte("<title>信頼性とログ先行書き込み</title>"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got1, got2 := titleMatch2(tt.args.src); !reflect.DeepEqual(got1, tt.want1) || !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("titleMatch2()1 = %v, want %v", string(got1), string(tt.want1))
				t.Errorf("titleMatch2()2  = %v, want %v", string(got2), string(tt.want2))
			}
		})
	}
}
