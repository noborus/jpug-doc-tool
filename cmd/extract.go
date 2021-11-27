package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type Pair struct {
	en string
	ja string
}

// コメント（英語原文）と続く文書（日本語翻訳）のペア、残り文字列、エラーを返す
// <!--
// english
// -->
// japanese
// の形式に一致しない場合はエラーを返す
func enjaPair(para []byte) (Pair, []byte, error) {
	re := EXCOMMENT.FindSubmatch(para)
	if len(re) < 3 {
		return Pair{}, nil, fmt.Errorf("no match")
	}
	enstr := strings.ReplaceAll(string(re[1]), "\n", " ")
	enstr = MultiSpace.ReplaceAllString(enstr, " ")
	enstr = strings.TrimSpace(enstr)

	jastr := strings.TrimSpace(string(re[2]))
	pair := Pair{
		en: enstr,
		ja: jastr,
	}

	if string(re[3]) == "<!--" {
		return pair, para[len(re[0])+3:], nil
	}

	return pair, nil, nil
}

// src を原文と日本語訳の対の配列に変換する
func Extraction(src []byte) []Pair {
	var pairs []Pair
	paras := REPARA.FindAll([]byte(src), -1)
	for _, para := range paras {
		pair, left, err := enjaPair(para)
		if err != nil {
			continue
		}
		pairs = append(pairs, pair)
		for len(left) > 0 {
			pair, left, err = enjaPair(left)
			if err != nil {
				continue
			}
			pairs = append(pairs, pair)
		}
	}

	rows := REROWS.FindAll([]byte(src), -1)
	for _, row := range rows {
		re := EXCOMMENT.FindSubmatch(row)
		if len(re) < 3 {
			continue
		}
		enstr := string(re[1])
		enstr = ENTRYSTRIP.ReplaceAllString(enstr, "")
		enstr = strings.ReplaceAll(enstr, "\n", " ")
		enstr = MultiSpace.ReplaceAllString(enstr, " ")
		enstr = strings.TrimSpace(enstr)

		jastr := string(re[2])
		jastr = ENTRYSTRIP.ReplaceAllString(jastr, "")
		jastr = strings.TrimSpace(jastr)

		pair := Pair{
			en: enstr,
			ja: jastr,
		}
		pairs = append(pairs, pair)
	}
	return pairs
}

func extract(fileNames []string) {
	for _, fileName := range fileNames {
		src, err := ReadFile(fileName)
		if err != nil {
			log.Fatal(err)
		}

		pairs := Extraction(src)

		dicname := DICDIR + fileName + ".t"
		f, err := os.Create(dicname)
		if err != nil {
			log.Fatal(err)
		}

		for _, pair := range pairs {
			fmt.Fprintf(f, "⦃%s⦀", pair.en)
			fmt.Fprintf(f, "%s⦄\n", pair.ja)
		}
		f.Close()
	}
}

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "英語と日本語翻訳を抽出する",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			extract(args)
		}
		fileNames := targetFileName()
		extract(fileNames)

	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
}
