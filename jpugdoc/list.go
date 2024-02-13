package jpugdoc

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jwalton/gchalk"
)

// 英日辞書の内容を表示する
// wf: ファイル名を表示する
// pre: preタグを表示する
// enOnly: 英文のみ表示する
// jaOnly: 日本語のみ表示する
// fileNames: ファイル名のリスト
func List(wf bool, pre bool, enOnly bool, jaOnly bool, fileNames []string) {
	w := io.Writer(os.Stdout)
	list(w, wf, pre, enOnly, jaOnly, fileNames)
}

func list(w io.Writer, wf bool, pre bool, enOnly bool, jaOnly bool, fileNames []string) {
	for _, fileName := range fileNames {
		if wf {
			fmt.Fprintln(w, gchalk.Red(fileName))
		}
		catalogs, err := loadCatalog(fileName)
		if err != nil {
			log.Println(err)
		}
		writeCatalog(w, catalogs, pre, enOnly, jaOnly)
	}
}

func writeCatalog(w io.Writer, catalogs []Catalog, suffix bool, enOnly bool, jaOnly bool) {
	for _, catalog := range catalogs {
		en := strings.Trim(catalog.en, "\n")
		if !suffix && en == "" {
			continue
		}

		if suffix {
			fmt.Fprintln(w, gchalk.Blue(strings.Trim(catalog.pre, "\n")))
		}
		if !jaOnly {
			fmt.Fprintln(w, gchalk.Green(en))
		}
		if !enOnly {
			fmt.Fprintln(w, strings.Trim(catalog.ja, "\n"))
		}
		if suffix {
			fmt.Fprintln(w, gchalk.Blue(strings.Trim(catalog.post, "\n")))
		}
		fmt.Fprintln(w)
	}
}

// 英日辞書の内容をTSV形式で表示する
func TSVList(fileNames []string) {
	w := io.Writer(os.Stdout)
	tsvList(w, fileNames)
}

func tsvList(w io.Writer, fileNames []string) {
	for _, fileName := range fileNames {
		catalogs, err := loadCatalog(fileName)
		if err != nil {
			log.Println(err)
		}
		writeTSV(w, catalogs)
	}
}

func writeTSV(w io.Writer, catalogs []Catalog) {
	for _, catalog := range catalogs {
		fmt.Fprintf(w, "%s\t%s\n", strings.ReplaceAll(catalog.en, "\n", " "), strings.ReplaceAll(catalog.ja, "\n", " "))
	}
}
