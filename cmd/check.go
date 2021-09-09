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
				fmt.Fprintln(out, "<========================================")
				fmt.Fprintln(out, gchalk.Red("コメントが始まっていません"))
				fmt.Fprintf(out, string(para))
				fmt.Fprintln(out)
				fmt.Fprintln(out, "========================================>")
				continue
			}
			// <literal>.*</literal>又は<literal>.*</literal><returnvalue>.*</returnvalue>又は+programlisting のみのparaだった場合は無視する
			if RELITERAL.Match(para) || RELIRET.Match(para) || RELIRETPROG.Match(para) {
				continue
			}
			fmt.Fprintln(out, "<========================================")
			fmt.Fprintln(out, gchalk.Red("コメントがありません"))
			fmt.Fprintln(out, string(para))
			fmt.Fprintln(out)
			fmt.Fprintln(out, "========================================>")
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

// enWordCheck は日本語翻訳中にある英単語が英語に含まれているかをチェックする
func enWordCheck(src []byte, word bool, tag bool, num bool) string {
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
				fmt.Fprintln(out, "<========================================")
				fmt.Fprintln(out, fmt.Sprintf("[%s]が含まれていません", gchalk.Red(strings.Join(unword, " ｜ "))))
				fmt.Fprintln(out, en)
				fmt.Fprintln(out, "-----------------------------------------")
				fmt.Fprintln(out, ja)
				fmt.Fprintln(out, "========================================>")
				fmt.Fprintln(out)
			}
		}

		if tag {
			untag := tagCheck(en, ja)
			if len(untag) > 0 {
				fmt.Fprintln(out, "<========================================")
				fmt.Fprintln(out, fmt.Sprintf("原文にあるタグ[%s]が含まれていません", gchalk.Red(strings.Join(untag, " ｜ "))))
				fmt.Fprintln(out, en)
				fmt.Fprintln(out, "-----------------------------------------")
				fmt.Fprintln(out, ja)
				fmt.Fprintln(out, "========================================>")
				fmt.Fprintln(out)
			}
		}

		if num {
			unNum := numCheck(en, ja)
			if len(unNum) > 0 {
				fmt.Fprintln(out, "<========================================")
				fmt.Fprintln(out, fmt.Sprintf("原文にある[%s]が含まれていません", gchalk.Red(strings.Join(unNum, " ｜ "))))
				fmt.Fprintln(out, en)
				fmt.Fprintln(out, "-----------------------------------------")
				fmt.Fprintln(out, ja)
				fmt.Fprintln(out, "========================================>")
				fmt.Fprintln(out)
			}
		}
	}
	return out.String()
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

		wCheck = enWordCheck(src, word, tag, num)
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
