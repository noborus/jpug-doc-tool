package jpugdoc

import (
	"fmt"
	"os"
	"strings"

	"github.com/jwalton/gchalk"
)

// CheckWord は英単語に対して日本語の単語が対になっているかをチェックします。
func CheckWord(en string, ja string, vTag string, fileNames []string) error {
	if vTag == "" {
		v, err := versionTag()
		if err != nil {
			return err
		}
		vTag = v
	}
	w := os.Stdout
	for _, fileName := range fileNames {
		src, err := getDiff(vTag, fileName)
		if err != nil {
			continue
		}
		pairs := Extraction(src)
		for _, pair := range pairs {
			s := strings.ReplaceAll(pair.en, "\n", " ")
			s = strings.Join(strings.Fields(s), " ")
			if en == "" || strings.Contains(s, en) {
				if ja == "" || (!strings.Contains(pair.ja, ja) && !strings.Contains(pair.ja, en)) {
					fmt.Println(gchalk.WithBgYellow().Black(fileName))
					fmt.Fprintln(w, gchalk.Green(pair.en))
					fmt.Fprintln(w, pair.ja)
				}
			}
		}
	}
	return nil
}
