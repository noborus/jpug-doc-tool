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
		keys := catalog.Keys()
		for _, v := range keys {
			val, ok := catalog.Get(v)
			if ok && !jaonly {
				fmt.Println(v)
			}
			if !enonly {
				fmt.Println(gchalk.Green(val.(string)))
			}
			fmt.Println()
		}
	}
}

func TSVList(fileNames []string) {
	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)
		keys := catalog.Keys()
		for _, en := range keys {
			ja, _ := catalog.Get(en)
			fmt.Printf("%s\t%s\n", strings.ReplaceAll(en, "\n", " "), strings.ReplaceAll(ja.(string), "\n", " "))
		}
	}
}
