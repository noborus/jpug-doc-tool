package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/jwalton/gchalk"
	"github.com/spf13/cobra"
	"golang.org/x/text/width"
)

type result struct {
	comment string
	en      string
	ja      string
}

type IgnoreList map[string]bool

func loadIgnore(fileName string) IgnoreList {
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	ignores := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ignores[scanner.Text()] = true
	}
	return ignores
}

// commentCheck は<para>内にコメント（<!-- -->)が含まれているかチェックする
func commentCheck(src []byte) []result {
	var results []result
	preComment := false
	for _, para := range REPARA.FindAll(src, -1) {
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
		if t == "<programlisting>" || t == "<screen>" || t == "<footnote>" || t == "<synopsis>" || t == "<replaceable>" || t == "</para>" {
			break
		}
		if !strings.Contains(ja, t) {
			unTag = append(unTag, t)
		}
	}
	return unTag
}

// 原文内の数値が日本語内にあるかチェックする
func numCheck(en string, ja string) []string {
	en = STRIPPROGRAMLISTING.ReplaceAllString(en, "")
	en = STRIPNUM.ReplaceAllString(en, "")
	ja = STRIPNUM.ReplaceAllString(ja, "")
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
func fileCheck(fileName string, src []byte, word bool, tag bool, num bool) []result {
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

		if word {
			unword := wordCheck(en, ja)
			if len(unword) > 0 {
				r := makeResult(fmt.Sprintf("[%s]が含まれていません", gchalk.Red(strings.Join(unword, " ｜ "))), en, ja)
				results = append(results, r)
			}
		}

		if tag {
			untag := tagCheck(en, ja)
			if len(untag) > 0 {
				r := makeResult(fmt.Sprintf("原文にある[%s]が含まれていません", gchalk.Red(strings.Join(untag, " ｜ "))), en, ja)
				results = append(results, r)
			}
		}

		if num {
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

func check(fileNames []string, ignore bool, word bool, tag bool, num bool) {
	for _, fileName := range fileNames {
		var results []result
		var ignores []string
		src, err := ReadFile(fileName)
		if err != nil {
			log.Fatal(err)
		}

		results = fileCheck(fileName, src, word, tag, num)
		if !word && !tag && !num {
			results = commentCheck(src)
		}

		if len(results) > 0 {
			fmt.Println(gchalk.Green(fileName))
			for _, r := range results {
				printResult(r)
				if ignore {
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
}

func registerIgnore(fileName string, ignores []string) {
	ignoreName := DICDIR + fileName + ".ignore"

	f, err := os.OpenFile(ignoreName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	for _, ig := range ignores {
		fmt.Fprintf(f, "%s\n", ig)
	}
}

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "文書をチェックする",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var word bool
		var tag bool
		var num bool
		var ignore bool
		var err error
		if word, err = cmd.PersistentFlags().GetBool("word"); err != nil {
			log.Println(err)
			return
		}
		if tag, err = cmd.PersistentFlags().GetBool("tag"); err != nil {
			log.Println(err)
			return
		}
		if num, err = cmd.PersistentFlags().GetBool("num"); err != nil {
			log.Println(err)
			return
		}
		if ignore, err = cmd.PersistentFlags().GetBool("ignore"); err != nil {
			log.Println(err)
			return
		}
		fileNames := targetFileName()
		if len(args) > 0 {
			fileNames = args
		}

		check(fileNames, ignore, word, tag, num)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.PersistentFlags().BoolP("word", "w", false, "Word check")
	checkCmd.PersistentFlags().BoolP("tag", "t", false, "Tag check")
	checkCmd.PersistentFlags().BoolP("num", "n", false, "Num check")
	checkCmd.PersistentFlags().BoolP("ignore", "i", false, "Prompt before ignore registration")
}
