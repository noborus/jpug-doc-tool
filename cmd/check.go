package cmd

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jwalton/gchalk"
	"github.com/spf13/cobra"
	"golang.org/x/text/width"
)

// commentCheck は<para>内にコメント（<!-- -->)が含まれているかチェックする
func commentCheck(src []byte) string {
	out := new(bytes.Buffer)
	for _, para := range REPARA.FindAll(src, -1) {
		if !containComment(para) {
			if containCommentEnd(para) {
				checkOutput(out, gchalk.Red("コメントが始まっていません"), string(para), "")
				continue
			}
			// <literal>.*</literal>又は<literal>.*</literal><returnvalue>.*</returnvalue>又は+programlisting のみのparaだった場合は無視する
			if RELITERAL.Match(para) || RELIRET.Match(para) || RELIRETPROG.Match(para) {
				continue
			}
			checkOutput(out, gchalk.Red("コメントがありません"), string(para), "")
		}
	}
	return out.String()
}

// 原文内のタグが日本語内にあるかチェックする
func tagCheck(en string, ja string) []string {
	tags := XMLTAG.FindAllString(en, -1)
	unTag := make([]string, 0)
	for _, t := range tags {
		if t == "<programlisting>" || t == "<screen>" || t == "<footnote>" || t == "<synopsis>" {
			continue
		}
		if !strings.Contains(strings.ToLower(ja), strings.ToLower(t)) {
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
func fileCheck(src []byte, word bool, tag bool, num bool) string {
	out := new(bytes.Buffer)
	for _, pair := range Extraction(src) {
		en := pair.en
		ja := pair.ja
		if len(ja) == 0 {
			continue
		}

		if word {
			unword := wordCheck(en, ja)
			if len(unword) > 0 {
				checkOutput(out, fmt.Sprintf("[%s]が含まれていません", gchalk.Red(strings.Join(unword, " ｜ "))), en, ja)
			}
		}

		if tag {
			untag := tagCheck(en, ja)
			if len(untag) > 0 {
				checkOutput(out, fmt.Sprintf("原文にある[%s]が含まれていません", gchalk.Red(strings.Join(untag, " ｜ "))), en, ja)
			}
		}

		if num {
			unNum := numCheck(en, ja)
			if len(unNum) > 0 {
				checkOutput(out, fmt.Sprintf("原文にある[%s]が含まれていません", gchalk.Red(strings.Join(unNum, " ｜ "))), en, ja)
			}
		}
	}
	return out.String()
}

// メッセージ、原文、日本語の形式で出力する
func checkOutput(out *bytes.Buffer, str string, en string, ja string) {
	fmt.Fprintln(out, "<========================================")
	fmt.Fprintln(out, str)
	fmt.Fprintln(out, gchalk.Green(en))
	fmt.Fprintln(out, "-----------------------------------------")
	fmt.Fprintln(out, ja)
	fmt.Fprintln(out, "========================================>")
	fmt.Fprintln(out)
}

func check(fileNames []string, word bool, tag bool, num bool) string {
	out := new(bytes.Buffer)
	for _, fileName := range fileNames {
		wCheck := ""
		cCheck := ""
		src, err := ReadFile(fileName)
		if err != nil {
			log.Fatal(err)
		}

		wCheck = fileCheck(src, word, tag, num)
		if !word && !tag && !num {
			cCheck = commentCheck(src)
		}

		if len(wCheck) > 0 || len(cCheck) > 0 {
			fmt.Fprintln(out, gchalk.Green(fileName))
		}
		if len(wCheck) > 0 {
			fmt.Fprintln(out, wCheck)
		}
		if len(cCheck) > 0 {
			fmt.Fprintln(out, cCheck)
		}
	}
	return out.String()
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
		fileNames := targetFileName()
		if len(args) > 0 {
			fileNames = args
		}

		out := check(fileNames, word, tag, num)
		fmt.Println(out)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.PersistentFlags().BoolP("word", "w", false, "Word check")
	checkCmd.PersistentFlags().BoolP("tag", "t", false, "Tag check")
	checkCmd.PersistentFlags().BoolP("num", "n", false, "Num check")
}
