package jpugdoc

import (
	"regexp"
)

// <para> </para> に一致させる
var REPARA = regexp.MustCompile(`(?s)(<para>\n*)(.*?)(\s*</para>)`)

// タグと一致させる
var RETAG = regexp.MustCompile(`<[^>]+>`)

// 単独タグ(\s+<programlisting>\nなど)に一致させる
var RETAGBLOCK = regexp.MustCompile(`^\s*<([^>]+)>\n`)

func regTagBlock(src string) []string {
	return RETAGBLOCK.FindAllString(src, -1)
}

var REPARA2 = regexp.MustCompile(`(?s)<para>\n(.*?)</para>`)

var CLOSEPARA = regexp.MustCompile(`</para>`)

// <para> <screen>|<programlisting> に一致させる
var REPARASCREEN = regexp.MustCompile(`(?s)(<para>\n*)(.*?)(\s*(</para>|<screen>|<programlisting>))`)

func regParaScreen(src []byte) [][]byte {
	return REPARASCREEN.FindAll(src, -1)
}

var BLANKBRACKET = regexp.MustCompile(`(?s)(<\!--\n(.*?)-->\n(.*?))(《》)\n`)

func similarBlank(src []byte) [][]byte {
	return BLANKBRACKET.FindAll(src, -1)
}

var STRIPPMT = regexp.MustCompile(`(?s)《機械翻訳》.*`)

// 《》で囲まれた文字列（作業中）に一致させる
var STRIPM = regexp.MustCompile(`《.*》`)

var BLANKBRACKET2 = regexp.MustCompile(`(?s)(.*?)-->\n(.*?)(《》)\n`)

// <para> </para> に一致させる（チェック用）
var CHECKPARA = regexp.MustCompile(`\s*(<!--\n?\s+)?(?s)(<para>\n*)(.*?)(\s*</para>)(\n*\s*-->)?`)

// <title> </title> に一致させる
var RETITLE = regexp.MustCompile(`(?s)(<title>\n*)(.*?)(\s*</title>)`)

// <title> </title> をコメントを含めて一致させる
var RECHECKTITLE = regexp.MustCompile(`(?s)(<!--\n\s*<title>\n*)(.*?)(\s*</title>\n*\s*-->\n\s*<title>(.*?)</title>)`)

var BLANKLINE = regexp.MustCompile(`^\s*\n`)

var BLANKSLINE = regexp.MustCompile(`(?m)^\n+`)

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
var REJASTRING = regexp.MustCompile(`[ぁ-ん]+|[ァ-ヴー]+|[一-龠]/`)

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

var (
	COMMENTSTART = regexp.MustCompile(`<!--`)
	COMMENTEND   = regexp.MustCompile(`-->`)
)

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

// コメント始まりの追加に一致
var STARTADDCOMMENT = regexp.MustCompile(`\+<!--$`)

// コメント終わりの追加に一致
var ENDADDCOMMENT = regexp.MustCompile(`\+-->$`)

// CDATAが終わりコメントの始まりに一致
var STARTADDCOMMENTWITHC = regexp.MustCompile(`\+(.*)?\]\]><!--`)

// コメントが終わりCDATAの始まりに一致
var ENDADDCOMMENTWITHC = regexp.MustCompile(`\+--><\!\[CDATA\[`)

// indexterm始まりに一致
var STARTINDEXTERM = regexp.MustCompile(`<indexterm`)

// indexterm終わりに一致
var ENDINDEXTERM = regexp.MustCompile(`</indexterm`)

// split-*コメントに一致
var SPLITCOMMENT = regexp.MustCompile(`split-.*-[start|end]`)

// コメント（英語原文）と続く文書（日本語翻訳）を取得
// 100%一致する訳ではない
var EXCOMMENT = regexp.MustCompile(`(?s)<!--(.*?)-->(.*?)(</row>|<!--|<note>|<informaltable>|<footnote>|<screen>|<synopsis>|<variablelist>|<programlisting>|<itemizedlist>|<simplelist>|<itemizedlist|<orderedlist|</para>)`)

func splitComment(src []byte) (en []byte, ja []byte, ex []byte) {
	re := EXCOMMENT.FindSubmatch(src)
	return re[1], re[2], re[3]
}

// 複数のスペースと一致
var (
	MultiSpace = regexp.MustCompile(`\s+`)
	MultiNL    = regexp.MustCompile(`\n+`)
)

// カタログから英語と日本語を取得
var SPLITCATALOG = regexp.MustCompile(`(?s)␝(.*?)␟(.*?)␟(.*?)␞(.*?)␞(.*?)␞`)

// 英単語 + /
var ENWORD = regexp.MustCompile(`[/a-zA-Z_]+`)

// 日本語
var NIHONGO = regexp.MustCompile(`[\p{Han}\p{Katakana}\p{Hiragana}]`)

// XMLタグ
var XMLTAG = regexp.MustCompile(`<[^>\!]*?>|<[^<>]+/>`)

// xreflabelは除外
var STRIPXREFLABEL = regexp.MustCompile(`xreflabel=\".*\"`)

// 数値
var ENNUM = regexp.MustCompile(`[0-9,]+`)

// <programlisting>.*</programlisting> を除外
var STRIPPROGRAMLISTING = regexp.MustCompile(`<(programlisting|screen)>.*</(programlisting|screen)>`)

func stripPROGRAMLISTING(src []byte) []byte {
	return STRIPPROGRAMLISTING.ReplaceAll(src, []byte(""))
}

// カンマを除外
var STRIPNUM = regexp.MustCompile(`,`)

// カンマを除外
var STRIPNUMJ = regexp.MustCompile(`, |,|、`)

// カタカナ
var KATAKANA = regexp.MustCompile(`[ァ-ヺー・]+`)

// 最後尾の日本語以外を除外
var STRIPNONJA = regexp.MustCompile(`[\s\,\(\)\.a-zA-Z0-9\-\/\<\>\n*]+$`)

func stripNONJA(src []byte) []byte {
	return STRIPNONJA.ReplaceAll(src, []byte(""))
}

// 文の終わりではないピリオドを除外
var NONSTRIPPERIOD = regexp.MustCompile(`\.\.\.|etc\.|Eg\.`)

func stripNONPERIOD(src string) string {
	return NONSTRIPPERIOD.ReplaceAllString(src, "")
}

var (
	KUTEN  = regexp.MustCompile(`。`)
	KUTEN2 = regexp.MustCompile(`。）? ?`)
)

var REPHIGHHUN = regexp.MustCompile(`&#0?45;`)

var REVHIGHHUN = regexp.MustCompile(`--`)

var REVHIGHHUN2 = regexp.MustCompile(`---`)

// 訳注を除外
var YAKUCHU = regexp.MustCompile(`[\(|\[|（]訳注[^\[|\)|）]]*[\]|\)|）]`)

var ENAUTHOR = regexp.MustCompile(`\([a-zA-ZÀ-ÿ, \.\-]+\)$`)

// 最後の作者(name)に一致させる
func authorMatch(src []byte) []byte {
	en := stripNL(string(src))
	ret := ENAUTHOR.FindAllString(en, -1)
	if len(ret) == 0 {
		return nil
	}
	return []byte(ret[len(ret)-1])
}

var TITLE = regexp.MustCompile(`(?U)\n( *<title>[a-zA-Z0-9 \.\-:]+</title>)\n([^\-])`)

func titleMatch(src []byte) []byte {
	ret := TITLE.FindSubmatch(src)
	if len(ret) == 0 {
		return nil
	}
	return ret[1]
}

var TITLE2 = regexp.MustCompile(`(?U)<!--\s*(<title>.*</title>)\s*-->\s*(<title>.*</title>)`)

func titleMatch2(src []byte) ([]byte, []byte) {
	ret := TITLE2.FindSubmatch(src)
	if len(ret) == 0 {
		return nil, nil
	}
	return ret[1], ret[2]
}

var RELEASENUM = regexp.MustCompile(`[0-9\.]+`)
