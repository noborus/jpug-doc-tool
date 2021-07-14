package cmd

import (
	"regexp"
)

// <para> </para> に一致させる
var REPARA = regexp.MustCompile(`(?s)(\<!--\n\s*)*(<para>\n*)(.*?)(\s*</para>)`)

// 文書から <para> </para>を取得してsliceで返す
func paraAll(src []byte) [][]byte {
	var ret [][]byte
	for _, para := range REPARA.FindAll(src, -1) {
		ret = append(ret, para)
	}
	return ret
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
func rowsAll(src []byte) [][]byte {
	var ret [][]byte
	for _, para := range REROWS.FindAll(src, -1) {
		ret = append(ret, para)
	}
	return ret
}

// <entry> </entry> と一致
var ENTRYSTRIP = regexp.MustCompile(`</?entry>`)

// entryタグを取り除くために使用
func stripEntry(src []byte) []byte {
	return ENTRYSTRIP.ReplaceAll(src, []byte(""))
}

// XMLのコメントに一致
var RECOMMENT = regexp.MustCompile(`(?s)<!--(.*?)-->`)

// コメントが含まれていれば true
func containComment(src []byte) bool {
	if RECOMMENT.Match(src) {
		return true
	}
	return false
}

// XMLのコメント閉じタグに一致
var RECOMMENTEND = regexp.MustCompile(`-->`)

// 先にコメントが含まれているかチェックして、コメント閉じタグのみの場合に true
func containCommentEnd(src []byte) bool {
	if RECOMMENTEND.Match(src) {
		return true
	}
	return false
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

// カタログから英語と日本語を取得
var SPLITCATALOG = regexp.MustCompile(`(?s)⦃(.*?)⦀(.*?)⦄`)

// 英単語 + /
var ENWORD = regexp.MustCompile(`[/a-zA-Z]+`)

// 数値
var ENNUM = regexp.MustCompile(`[0-9]+`)

// カタカナ
var KATAKANA = regexp.MustCompile(`[ァ-ヺー・]+`)

// 最後尾の日本語以外を除外
var STRIPNONJA = regexp.MustCompile(`[\s\,\(\)\.a-zA-Z0-9\-\/\<\>\n*]+$`)

func stripNONJA(src []byte) []byte {
	return STRIPNONJA.ReplaceAll(src, []byte(""))
}
