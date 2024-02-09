package jpugdoc

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jwalton/gchalk"
)

// 英日辞書の内容を表示する
// wf: ファイル名を表示する
// affix: pre/postを表示する
// enOnly: 英文のみ表示する
// jaOnly: 日本語のみ表示する
// fileNames: ファイル名のリスト
func List(wf bool, affix bool, enOnly bool, jaOnly bool, fileNames []string) {
	list(os.Stdout, wf, affix, enOnly, jaOnly, fileNames)
}

func list(w io.Writer, wf bool, affix bool, enOnly bool, jaOnly bool, fileNames []string) {
	for _, fileName := range fileNames {
		if wf {
			fmt.Fprintln(w, gchalk.Red(fileName))
		}
		catalogs := loadCatalog(fileName)
		writeCatalog(w, catalogs, affix, enOnly, jaOnly)
	}
}

func writeCatalog(w io.Writer, catalogs []Catalog, affix bool, enOnly bool, jaOnly bool) {
	for _, catalog := range catalogs {
		if affix {
			fmt.Fprintln(w, gchalk.Blue(catalog.pre))
		} else {
			if catalog.en == "" {
				continue
			}
		}
		if !jaOnly {
			fmt.Fprintln(w, gchalk.Green(catalog.en))
		}
		if !enOnly {
			fmt.Fprintln(w, catalog.ja)
		}
		if affix {
			fmt.Fprintln(w, gchalk.Blue(catalog.post))
		}
		fmt.Fprintln(w)
	}
}

// 英日辞書の内容をTSV形式で表示する
func TSVList(fileNames []string) {
	tsvList(os.Stdout, fileNames)
}

func tsvList(w io.Writer, fileNames []string) {
	for _, fileName := range fileNames {
		catalogs := loadCatalog(fileName)
		writeTSV(w, catalogs)
	}
}

func writeTSV(w io.Writer, catalogs []Catalog) {
	for _, catalog := range catalogs {
		fmt.Fprintf(w, "%s\t%s\n", strings.ReplaceAll(catalog.en, "\n", " "), strings.ReplaceAll(catalog.ja, "\n", " "))
	}
}
