package jpugdoc

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/jwalton/gchalk"
	"golang.org/x/text/width"
)

type result struct {
	comment string
	en      string
	ja      string
}

// CheckFlag represents the item to check.
type CheckFlag struct {
	VTag   string
	Ignore bool
	WIP    bool
	Para   bool
	Word   bool
	Tag    bool
	Num    bool
	Strict bool
}

// default check
func Check(fileNames []string, cf CheckFlag) error {
	if cf.VTag == "" {
		v, err := versionTag()
		if err != nil {
			return err
		}
		cf.VTag = v
	}

	for _, fileName := range fileNames {
		if cf.Para {
			fileCheck(fileName, cf)
		}

		diffSrc, err := getDiff(cf.VTag, fileName)
		if err != nil {
			continue
		}
		formatCheck(fileName, diffSrc, cf)
		translationCheck(fileName, diffSrc, cf)
	}
	return nil
}

// メッセージ、原文、日本語の形式で出力する
func makeResult(str string, en string, ja string) result {
	var r result
	r.comment = str
	r.en = en
	r.ja = ja
	return r
}

func printResult(r result) {
	fmt.Println("<========================================")
	fmt.Println(r.comment)
	fmt.Println(gchalk.Green(r.en))
	if r.ja != "" {
		fmt.Println("-----------------------------------------")
		fmt.Println(r.ja)
	}
	fmt.Println("========================================>")
}

// enjaCheck は英語と日本語翻訳から wordCheck,numOfTagCheck, numCheckをチェックする
func enjaCheck(fileName string, catalog Catalog, cf CheckFlag) []result {
	var results []result
	en := catalog.en
	en = MultiNL.ReplaceAllString(en, " ")
	en = MultiSpace.ReplaceAllString(en, " ")
	ja := catalog.ja
	ja = MultiNL.ReplaceAllString(ja, " ")
	ja = MultiSpace.ReplaceAllString(ja, " ")
	ja = YAKUCHU.ReplaceAllString(ja, "")

	if en == "" || ja == "" {
		return nil
	}
	// 作業中(cf.WIPがfalse)の箇所は"《"が含まれている場合はチェックしない
	if !cf.WIP {
		if strings.Contains(ja, "《") {
			return nil
		}
	}

	if cf.Word {
		unWord := wordCheck(en, ja)
		if len(unWord) > 0 {
			r := makeResult(fmt.Sprintf("[%s]が含まれていません", gchalk.Red(strings.Join(unWord, " ｜ "))), en, ja)
			results = append(results, r)
		}
	}

	if cf.Tag {
		numTag := numOfTagCheck(cf.Strict, en, ja)
		if len(numTag) > 0 {
			r := makeResult(fmt.Sprintf("タグ[%s]の数が違います", gchalk.Red(strings.Join(numTag, " ｜ "))), en, ja)
			results = append(results, r)
		}
	}

	if cf.Num {
		unNum := numCheck(en, ja)
		if len(unNum) > 0 {
			r := makeResult(fmt.Sprintf("原文にある[%s]が含まれていません", gchalk.Red(strings.Join(unNum, " ｜ "))), en, ja)
			results = append(results, r)
		}
	}

	return results
}

// 英語と日本語のタグの数をチェックする
func numOfTagCheck(strict bool, en string, ja string) []string {
	tags := XMLTAG.FindAllString(en, -1)
	unTag := make([]string, 0)
	for _, t := range tags {
		t = STRIPXREFLABEL.ReplaceAllString(t, "")
		if strict {
			if strings.Count(en, t) != strings.Count(ja, t) {
				unTag = append(unTag, fmt.Sprintf("(%s)%d:%d", t, strings.Count(en, t), strings.Count(ja, t)))
			}
		} else {
			if strings.Count(en, t) > strings.Count(ja, t) {
				unTag = append(unTag, fmt.Sprintf("(%s)%d:%d", t, strings.Count(en, t), strings.Count(ja, t)))
			}
		}

	}
	return unTag
}

// 原文内の数値が日本語内にあるかチェックする
func numCheck(en string, ja string) []string {
	en = STRIPPROGRAMLISTING.ReplaceAllString(en, "")
	en = STRIPNUM.ReplaceAllString(en, "")
	ja = STRIPNUMJ.ReplaceAllString(ja, "")
	nums := ENNUM.FindAllString(en, -1)
	unNum := make([]string, 0)
	for _, n := range nums {
		// -- → &#045; は除外、zero -> 0も除外
		if n == "045" || n == "45" || n == "0" {
			continue
		}
		if !strings.Contains(width.Narrow.String(ja), n) {
			unNum = append(unNum, n)
		}
	}
	return unNum
}

// 日本語訳内の英単語、数字が原文に含まれているかチェックする
func wordCheck(en string, ja string) []string {
	ja = STRIPNONJA.ReplaceAllString(ja, "")
	words := ENWORD.FindAllString(ja, -1)
	unWord := make([]string, 0)
	for _, w := range words {
		if !strings.Contains(strings.ToLower(en), strings.ToLower(w)) {
			unWord = append(unWord, w)
		}
	}
	return unWord
}

// diffの内容から英日のブロックを抽出して整合をチェックする
func translationCheck(fileName string, src []byte, cf CheckFlag) {
	var ignores []string
	var results []result

	catalogs := Extraction(src)
	for _, c := range catalogs {
		result := enjaCheck(fileName, c, cf)
		if len(result) > 0 {
			results = append(results, result...)
		}
	}

	ignoreName := DicDir + fileName + ".ignore"
	ignoreList := loadIgnore(ignoreName)

	if len(results) > 0 {
		fmt.Println(gchalk.Green(fileName))
		for _, r := range results {
			if ignoreList[strings.TrimRight(r.en, "\n")] {
				continue
			}
			printResult(r)
			if cf.Ignore {
				if prompter.YN("ignore?", false) {
					ignores = append(ignores, r.en)
				}
			}
		}
	}
	if len(ignores) > 0 {
		registerIgnore(ignoreName, ignores)
	}
}

// git diff を取り内容をチェックする。
func formatCheck(fileName string, diffSrc []byte, cf CheckFlag) {
	var results []result
	var ignores []string

	ignoreName := DicDir + fileName + ".ignore"
	ignoreList := loadIgnore(ignoreName)

	results = commentCheck(diffSrc, cf)
	if len(results) > 0 {
		fmt.Println(gchalk.Green(fileName))
		for _, r := range results {
			en := stripNL(r.en)
			if ignoreList[en] {
				continue
			}
			printResult(r)
			if cf.Ignore {
				if prompter.YN("ignore?", false) {
					ignores = append(ignores, en)
				}
			}
		}
	}
	if len(ignores) > 0 {
		registerIgnore(ignoreName, ignores)
	}
}

// diffの内容から追加されたコメントの開始と終了をチェックする。
func commentCheck(diffSrc []byte, cf CheckFlag) []result {
	var results []result

	reader := bytes.NewReader(diffSrc)
	scanner := bufio.NewScanner(reader)
	var en, ja strings.Builder
	var comment bool

	// skip diff header
	skipHeader(scanner)

	for scanner.Scan() {
		l := scanner.Text()
		line := strings.TrimSpace(l)
		if STARTADDCOMMENT.MatchString(line) || STARTADDCOMMENTWITHC.MatchString(line) { // <!--
			if comment {
				r := makeResult(gchalk.Red("コメント位置が不正"), en.String(), ja.String())
				results = append(results, r)
			}
			comment = true
			ja.Reset()
			en.Reset()
		} else if ENDADDCOMMENT.MatchString(line) || ENDADDCOMMENTWITHC.MatchString(line) { // -->
			comment = false
		}
		if comment {
			en.WriteString(l[1:])
			en.WriteString("\n")
		} else if l[0] == '+' {
			ja.WriteString(l[1:])
			ja.WriteString("\n")
		}
	}
	return results
}

// ファイル自体チェックする
func fileCheck(fileName string, cf CheckFlag) error {
	var results []result
	var ignores []string

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	ignoreName := DicDir + fileName + ".ignore"
	//ignoreList := loadIgnore(ignoreName)
	var buf strings.Builder
	var paraFlag, commentFlag, ok bool
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		buf.WriteString(l)
		buf.WriteRune('\n')
		if !commentFlag && strings.Contains(l, "<para>") {
			paraFlag = true
			continue
		}
		if paraFlag && strings.Contains(l, "<!--") {
			commentFlag = true
			ok = true
		}

		if strings.Contains(l, "-->") {
			commentFlag = false
			continue
		}

		if !commentFlag && strings.Contains(l, "</para>") {
			paraFlag = false
			ok = false
			buf.Reset()
			continue
		}
		if paraFlag && !ok {
			r := makeResult(gchalk.Red("コメントがありません"), buf.String(), "")
			results = append(results, r)
			buf.Reset()
		}
	}

	if len(results) > 0 {
		fmt.Println(gchalk.Green(fileName))
		for _, r := range results {
			printResult(r)
			if cf.Ignore {
				if prompter.YN("ignore?", false) {
					ignores = append(ignores, r.en)
				}
			}
		}
	}
	if len(ignores) > 0 {
		registerIgnore(ignoreName, ignores)
	}
	return nil
}

// paraCheck は<para>内にコメント（<!-- -->)が含まれているかチェックする
func paraCheck(src []byte) []result {
	var results []result
	p := 0
	pp := 0
	for p < len(src) {
		pp = bytes.Index(src[p:], []byte("<para>"))
		if pp == -1 {
			break
		}
		if inComment(src[:p+pp]) {
			p = p + pp + 7
			continue
		}
		e := bytes.Index(src[p+pp:], []byte("</para>"))
		if e == -1 {
			break
		}
		if bytes.Contains(src, []byte("<returnvalue>")) {
			p = p + pp + 7
			continue
		}
		if !bytes.Contains(src[p:p+pp+e], []byte("<!--")) {
			if !NIHONGO.Match(src[p+pp : p+pp+e+8]) {
				r := makeResult(gchalk.Red("コメントがありません"), string(src[p+pp:p+pp+e+8]), "")
				results = append(results, r)
			}
		}
		p = p + pp + e + 8
	}
	return results
}
