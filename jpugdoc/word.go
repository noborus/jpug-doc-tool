package jpugdoc

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jwalton/gchalk"
)

// CheckWord は英単語に対して日本語の単語が対になっているかをチェックする
// en: 英単語
// ja: 日本語の単語
// vTag: バージョンタグ
// fileNames: ファイル名
func CheckWord(en string, ja string, vTag string, fileNames []string) error {
	if vTag == "" {
		v, err := versionTag()
		if err != nil {
			return err
		}
		vTag = v
	}

	w := os.Stdout

	found := false
	for _, fileName := range fileNames {
		src, err := getDiff(vTag, fileName)
		if err != nil {
			continue
		}
		pairs := Extraction(src)
		for _, pair := range pairs {
			var eExist, jExist bool
			if en != "" {
				eExist = checkPairEn(pair, en)
			}
			if ja != "" {
				jExist = checkPairJa(pair, en, ja)
			}

			writeF := false
			if en != "" && ja == "" { // enのみ指定された場合
				writeF = eExist
			} else if en == "" && ja != "" { // jaのみ指定された場合
				writeF = jExist
			} else if en != "" && ja != "" { // en, jaが指定された場合は、英語が含まれていて日本語が含まれて*いない*ものを出力
				writeF = eExist && !jExist
			}

			if writeF {
				fmt.Println(gchalk.WithBgYellow().Black(fileName))
				fmt.Fprintln(w, gchalk.Green(pair.en))
				fmt.Fprintln(w, pair.ja)
				found = true
			}
		}
	}
	if !found {
		fmt.Fprintln(os.Stderr, gchalk.Red("Not found", ":", en))
	}
	return nil
}

func checkPairEn(pair Catalog, en string) bool {
	enStr := strings.ReplaceAll(pair.en, "\n", " ")
	enStr = strings.Join(strings.Fields(enStr), " ")
	if en == "" || strings.Contains(enStr, en) {
		return true
	}
	return false
}

func checkPairJa(pair Catalog, en string, ja string) bool {
	if ja == "" {
		return false
	}
	words := []string{}
	if en != "" {
		words = []string{en}
	}
	str, err := strconv.Unquote(ja)
	if err != nil {
		str = ja
	}
	words = append(words, strings.Split(str, ",")...)
	return conteinsWord(pair.ja, words)
}

func conteinsWord(s string, words []string) bool {
	for _, word := range words {
		word = strings.TrimSpace(word)
		if strings.Contains(s, word) {
			return true
		}
	}
	return false
}
