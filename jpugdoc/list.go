package jpugdoc

import (
	"fmt"
	"strings"

	"github.com/jwalton/gchalk"
)

// 英日辞書の内容を表示する
func List(wf bool, pre bool, enonly bool, jaonly bool, fileNames []string) {
	for _, fileName := range fileNames {
		if wf {
			fmt.Println(gchalk.Red(fileName))
		}
		catalog := loadCatalog(fileName)
		for _, c := range catalog {
			if pre {
				fmt.Println(gchalk.Blue(c.pre))
			} else {
				if c.en == "" {
					continue
				}
			}
			if !jaonly {
				fmt.Println(gchalk.Green(c.en))
			}
			if !enonly {
				fmt.Println(c.ja)
			}
			fmt.Println()
		}
	}
}

// 英日辞書の内容をTSV形式で表示する
func TSVList(fileNames []string) {
	for _, fileName := range fileNames {
		catalog := loadCatalog(fileName)
		for _, c := range catalog {
			fmt.Printf("%s\t%s\n", strings.ReplaceAll(c.en, "\n", " "), strings.ReplaceAll(c.ja, "\n", " "))
		}
	}
}
