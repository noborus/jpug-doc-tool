package jpugdoc

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jwalton/gchalk"
)

// CheckWord checks if the given English word (en) is paired with the corresponding Japanese word (ja) in the specified files.
// If both en and ja are provided, it identifies pairs where the English word exists but the Japanese word does not.
// If only en is provided, it checks for the presence of the English word.
// If only ja is provided, it checks for the presence of the Japanese word.
//
// Parameters:
//   - en: The English word to search for.
//   - ja: The Japanese word to search for.
//   - vTag: The version tag to use for retrieving file differences. If empty, it will be determined automatically.
//   - fileNames: A list of file names to search within.
//
// Returns:
//   - An error if there is an issue retrieving the version tag or file differences, or nil if the operation completes successfully.
//
// Example:
//
//	err := CheckWord("example", "例", "v1.0", []string{"file1.txt", "file2.txt"})
//	if err != nil {
//	    log.Fatal(err)
//	}
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
	eCount, jCount := 0, 0
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
				if eExist {
					eCount++
				}
			}
			if ja != "" {
				jExist = checkPairJa(pair, en, ja)
				if jExist {
					jCount++
				}
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
				fmt.Fprintln(w, gchalk.WithBgYellow().Black(fileName))
				fmt.Fprintln(w, gchalk.Green(pair.en))
				fmt.Fprintln(w, pair.ja)
				found = true
			}
		}
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "[%s](%d):[%s](%d)\n", en, eCount, ja, jCount)
	if !found && eCount == 0 && jCount == 0 {
		fmt.Fprintln(os.Stderr, gchalk.Red("Not found: "+en))
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
