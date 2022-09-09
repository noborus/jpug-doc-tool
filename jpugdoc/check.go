package jpugdoc

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
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
	Ignore bool
	Word   bool
	Tag    bool
	Num    bool
}

type IgnoreList map[string]bool

func loadIgnore(fileName string) IgnoreList {
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer f.Close()

	ignores := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ignores[scanner.Text()] = true
	}
	return ignores
}

func registerIgnore(fileName string, ignores []string) {
	ignoreName := DICDIR + fileName + ".ignore"

	f, err := os.OpenFile(ignoreName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	for _, ig := range ignores {
		fmt.Fprintf(f, "%s\n", ig)
	}
}

// commentCheck は<para>内にコメント（<!-- -->)が含まれているかチェックする
func commentCheck(src []byte) []result {
	var results []result
	preComment := false
	for _, para := range CHECKPARA.FindAll(src, -1) {
		para = BLANKLINE.ReplaceAll(para, []byte(""))
		if !containComment(para) {
			if containCommentEnd(para) {
				r := makeResult(gchalk.Red("コメントが始まっていません"), string(para), "")
				results = append(results, r)
				continue
			}
			// <literal>.*</literal>又は<literal>.*</literal><returnvalue>.*</returnvalue>又は+programlisting のみのparaだった場合は無視する
			if RELITERAL.Match(para) || RELIRET.Match(para) || RELIRETPROG.Match(para) {
				continue
			}
			if !preComment {
				r := makeResult(gchalk.Red("コメントがありません"), string(para), "")
				results = append(results, r)
			}
			preComment = false
			continue
		}
		if endComment(para) {
			preComment = true
		} else {
			preComment = false
		}
	}
	return results
}

// 原文内のタグが日本語内にあるかチェックする
func tagCheck(en string, ja string) []string {
	tags := XMLTAG.FindAllString(en, -1)
	unTag := make([]string, 0)
	for _, t := range tags {
		if !strings.Contains(ja, t) {
			unTag = append(unTag, t)
		}
	}
	return unTag
}

func numOfTagCheck(en string, ja string) []string {
	tags := XMLTAG.FindAllString(en, -1)
	unTag := make([]string, 0)
	for _, t := range tags {
		if strings.Count(en, t) > strings.Count(ja, t) {
			unTag = append(unTag, fmt.Sprintf("(%s)%d:%d", t, strings.Count(en, t), strings.Count(ja, t)))

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
		// &#045;
		if n == "045" || n == "45" || n == "0" {
			continue
		}
		if !strings.Contains(width.Narrow.String(ja), n) {
			unNum = append(unNum, n)
		}
	}
	return unNum
}

// 日本語訳内の英単語が原文に含まれているかチェックする
func wordCheck(en string, ja string) []string {
	ja = STRIPNONJA.ReplaceAllString(ja, "")
	words := ENWORD.FindAllString(ja, -1)
	num := ENNUM.FindAllString(ja, -1)
	for _, n := range num {
		i, err := strconv.Atoi(n)
		if err == nil || i < 5 {
			continue
		}
		words = append(words, n)
	}
	unword := make([]string, 0)
	for _, w := range words {
		if !strings.Contains(strings.ToLower(en), strings.ToLower(w)) {
			unword = append(unword, w)
		}
	}
	return unword
}

// fileCheck は日本語翻訳中にある英単語が英語に含まれているかをチェックする
func fileCheck(fileName string, src []byte, cf CheckFlag) []result {
	var results []result
	ignoreName := DICDIR + fileName + ".ignore"
	ignores := loadIgnore(ignoreName)
	for _, pair := range Extraction(src) {
		en := pair.en
		ja := pair.ja
		if len(en) == 0 || len(ja) == 0 {
			continue
		}
		if ignores[en] {
			continue
		}
		ja = MultiNL.ReplaceAllString(ja, " ")
		ja = MultiSpace.ReplaceAllString(ja, " ")

		if cf.Word {
			unword := wordCheck(en, ja)
			if len(unword) > 0 {
				r := makeResult(fmt.Sprintf("[%s]が含まれていません", gchalk.Red(strings.Join(unword, " ｜ "))), en, ja)
				results = append(results, r)
			}
		}

		if cf.Tag {
			numTag := numOfTagCheck(en, ja)
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
	}
	return results
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

func oldCheck(fileName string, cf CheckFlag) {
	var results []result
	var ignores []string
	src, err := ReadAllFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	results = fileCheck(fileName, src, cf)
	if !cf.Word && !cf.Tag && !cf.Num {
		results = commentCheck(src)
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
		registerIgnore(fileName, ignores)
	}
}

// enjaCheck は日本語翻訳中にある英単語が英語に含まれているかをチェックする
func enjaCheck(fileName string, catalog Catalog, cf CheckFlag) []result {
	var results []result
	en := catalog.en
	en = MultiNL.ReplaceAllString(en, " ")
	en = MultiSpace.ReplaceAllString(en, " ")
	ja := catalog.ja
	ja = MultiNL.ReplaceAllString(ja, " ")
	ja = MultiSpace.ReplaceAllString(ja, " ")
	if en == "" {
		return nil
	}
	if cf.Word {
		unword := wordCheck(en, ja)
		if len(unword) > 0 {
			r := makeResult(fmt.Sprintf("[%s]が含まれていません", gchalk.Red(strings.Join(unword, " ｜ "))), en, ja)
			results = append(results, r)
		}
	}

	if cf.Tag {
		numTag := numOfTagCheck(en, ja)
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

func listCheck(fileName string, src []byte, cf CheckFlag) {
	var ignores []string
	var results []result
	catalogs := Extraction(src)
	for _, c := range catalogs {
		result := enjaCheck(fileName, c, cf)
		if len(result) > 0 {
			results = append(results, result...)
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
		registerIgnore(fileName, ignores)
	}
}

func gitCheck(fileName string, src []byte, cf CheckFlag) {
	var results []result
	var ignores []string

	results = checkDiff(src)
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
		registerIgnore(fileName, ignores)
	}
}

func checkDiff(src []byte) []result {
	var results []result
	reader := bytes.NewReader(src)
	scanner := bufio.NewScanner(reader)
	var en, ja strings.Builder
	var comment bool
	for scanner.Scan() {
		l := scanner.Text()
		line := strings.TrimSpace(l)
		if STARTADDCOMMENT.MatchString(line) || STARTADDCOMMENTWITHC.MatchString(line) {
			if comment {
				r := makeResult(gchalk.Red("コメント位置が不正"), en.String(), ja.String())
				results = append(results, r)
			}
			comment = true
			ja.Reset()
			en.Reset()
		} else if ENDADDCOMMENT.MatchString(line) || ENDADDCOMMENTWITHC.MatchString(line) {
			comment = false
		}
		if comment {
			en.WriteString(l[1:])
			en.WriteString("\n")
		} else {
			ja.WriteString(l[1:])
			ja.WriteString("\n")
		}
	}
	return results
}

// Check
func Check(fileNames []string, cf CheckFlag) {
	vTag, err := versionTag()
	if err != nil {
		log.Fatal(err)
	}
	for _, fileName := range fileNames {

		src := getDiff(vTag, fileName)
		gitCheck(fileName, src, cf)
		listCheck(fileName, src, cf)
	}
}
