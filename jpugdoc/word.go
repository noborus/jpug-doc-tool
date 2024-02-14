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
			if strings.Contains(pair.en, en) {
				if !strings.Contains(pair.ja, ja) {
					fmt.Fprintf(w, "%s not include word)\n", fileName)
					fmt.Fprintln(w, gchalk.Green(pair.en))
					fmt.Fprintln(w, pair.ja)
				}
			}
		}
	}
	return nil
}
