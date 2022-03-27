package jpugdoc

import (
	"regexp"
)

// <para> </para> に一致させる
var REPARA = regexp.MustCompile(`(?s)(<para>\n*)(.*?)(\s*</para>)`)

// <para> </para> に一致させる（チェック用）
var CHECKPARA = regexp.MustCompile(`\s*(<!--\n?\s+)?(?s)(<para>\n*)(.*?)(\s*</para>)(\n*\s*-->)?`)

var BLANKLINE = regexp.MustCompile(`^\s*\n`)

// 文書から <para> </para>を取得してsliceで返す
func ParaAll(src []byte) [][]byte {
	return CHECKPARA.FindAll(src, -1)
}

// <para><literal></literal></para>に一致させる
var RELITERAL = regexp.MustCompile(`(?s)(<para>\n*)\s*<literal>.*</literal>\n*(\s*</para>)`)

// <para><literal></literal><returnvalue></returnvalue></para>に一致させる
var RELIRET = regexp.MustCompile(`(?s)(<para>\n*)\s*<literal>.*</literal>\n*\s*<returnvalue>.*</returnvalue>(\s*</para>)`)

// <para><literal></literal><returnvalue></returnvalue><programlisting></programlisting></para>に一致させる
var RELIRETPROG = regexp.MustCompile(`(?s)(<para>\n*)\s*<literal>.*</literal>\n*\s*<returnvalue>.*</returnvalue>\n*\s*<programlisting>.*</programlisting>(\s*</para>)`)

// var RELIRET = regexp.MustCompile(`(?s)(<para>\n*)\s*<literal>.*</literal>\n`)

// <row> </row> に一致させる
var REROWS = regexp.MustCompile(`(?s)(<row>\n*)(.*?)(\s+</row>)`)

// 文書から <rows> </rows>を取得してsliceで返す
func RowsAll(src []byte) [][]byte {
	return REROWS.FindAll(src, -1)
}

// <entry> </entry> と一致
var ENTRYSTRIP = regexp.MustCompile(`</?entry>`)

// entryタグを取り除くために使用
func StripEntry(src []byte) []byte {
	return ENTRYSTRIP.ReplaceAll(src, []byte(""))
}

// XMLのコメントに一致
var RECOMMENT = regexp.MustCompile(`(?s)<!--(.*?)-->`)

// コメントが含まれていれば true
func containComment(src []byte) bool {
	return RECOMMENT.Match(src)
}

// XMLのコメント開始タグに一致
var RECOMMENTSTART = regexp.MustCompile(`<!--`)

// XMLのコメント閉じタグに一致
var RECOMMENTEND = regexp.MustCompile(`-->`)

// 先にコメントが含まれているかチェックして、コメント閉じタグのみの場合に true
func containCommentEnd(src []byte) bool {
	return RECOMMENTEND.Match(src)
}

// 最初がコメント始まりに一致
var STARTCOMMENT = regexp.MustCompile(`^<!--`)

// 最後がコメント終わりに一致
var ENDCOMMENT = regexp.MustCompile(`-->$`)

// 最後がコメント終わりであればtrue
func endComment(src []byte) bool {
	return ENDCOMMENT.Match(src)
}

// コメント（英語原文）と続く文書（日本語翻訳）を取得
// 100%一致する訳ではない
var EXCOMMENT = regexp.MustCompile(`(?s)<!--(.*?)-->(.*?)(</row>|<!--|<note>|<informaltable>|<footnote>|<screen>|<synopsis>|<variablelist>|<programlisting>|<itemizedlist>|<simplelist>|<itemizedlist|<orderedlist|</para>)`)

func splitComment(src []byte) (en []byte, ja []byte, ex []byte) {
	re := EXCOMMENT.FindSubmatch(src)
	return re[1], re[2], re[3]
}

// 複数のスペースと一致
var MultiSpace = regexp.MustCompile(`\s+`)
var MultiNL = regexp.MustCompile(`\n+`)

// カタログから英語と日本語を取得
var SPLITCATALOG = regexp.MustCompile(`(?s)␝(.*?)␟(.*?)␞`)

// 英単語 + /
var ENWORD = regexp.MustCompile(`[/a-zA-Z_]+`)

// XMLタグ
var XMLTAG = regexp.MustCompile(`<[^>]*?>|<[^<>]+/>`)

// 数値
var ENNUM = regexp.MustCompile(`[0-9,]+`)

// <programlisting>.*</programlisting> を除外
var STRIPPROGRAMLISTING = regexp.MustCompile(`<(programlisting|screen)>.*</(programlisting|screen)>`)

func stripPROGRAMLISTING(src []byte) []byte {
	return STRIPPROGRAMLISTING.ReplaceAll(src, []byte(""))
}

// カンマを除外
var STRIPNUM = regexp.MustCompile(`,`)

// カタカナ
var KATAKANA = regexp.MustCompile(`[ァ-ヺー・]+`)

// 最後尾の日本語以外を除外
var STRIPNONJA = regexp.MustCompile(`[\s\,\(\)\.a-zA-Z0-9\-\/\<\>\n*]+$`)

func stripNONJA(src []byte) []byte {
	return STRIPNONJA.ReplaceAll(src, []byte(""))
}

var KUTEN = regexp.MustCompile(`。`)
