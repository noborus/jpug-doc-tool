package jpugdoc

import (
	"fmt"
	"strings"

	"github.com/jwalton/gchalk"
)

func List(enonly bool, jaonly bool, fileNames []string) {
	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)
		for en, ja := range catalog {
			if !jaonly {
				fmt.Println(gchalk.Green(en))
			}
			if !enonly {
				fmt.Println(ja)
			}
			fmt.Println()
		}
	}
}

func TSVList(fileNames []string) {
	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)
		for en, ja := range catalog {
			fmt.Printf("%s\t%s\n", strings.ReplaceAll(en, "\n", " "), strings.ReplaceAll(ja, "\n", " "))
		}
	}
}
