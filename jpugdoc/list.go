package jpugdoc

import (
	"fmt"
	"strings"

	"github.com/jwalton/gchalk"
)

func List(wf bool, pre bool, enonly bool, jaonly bool, fileNames []string) {
	for _, fileName := range fileNames {
		if wf {
			fmt.Println(gchalk.Red(fileName))
		}
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)
		for _, c := range catalog {
			if pre {
				fmt.Println(gchalk.Blue(c.pre))
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

func TSVList(fileNames []string) {
	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)
		for _, c := range catalog {
			fmt.Printf("%s\t%s\n", strings.ReplaceAll(c.en, "\n", " "), strings.ReplaceAll(c.ja, "\n", " "))
		}
	}
}
