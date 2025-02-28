package jpugdoc

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jwalton/gchalk"
)

type ListOoptions struct {
	WriteFile bool // wf: ファイル名を表示する
	IsPre     bool // pre: preタグを表示する
	ENOnly    bool // enOnly: 英文のみ表示する
	JAOnly    bool // jaOnly: 日本語のみ表示する
	Strip     bool // strip: 空白を削除する
	Sentence  bool // sentence: 文で改行する
}

// 英日辞書の内容を表示する
// fileNames: ファイル名のリスト
func List(opt ListOoptions, fileNames []string) {
	w := io.Writer(os.Stdout)
	list(w, opt, fileNames)
}

func list(w io.Writer, opt ListOoptions, fileNames []string) {
	for _, fileName := range fileNames {
		if opt.WriteFile {
			fmt.Fprintln(w, gchalk.Red(fileName))
		}
		catalogs, err := loadCatalog(fileName)
		if err != nil {
			log.Println(err)
		}
		writeCatalog(w, catalogs, opt)
	}
}

func ListCommon(opt ListOoptions) {
	w := io.Writer(os.Stdout)
	listCommon(w, opt)
}

func listCommon(w io.Writer, opt ListOoptions) {
	catalogs, err := loadCatalog("common")
	if err != nil {
		log.Println(err)
	}
	writeCatalog(w, catalogs, opt)
}

func writeCatalog(w io.Writer, catalogs []Catalog, opt ListOoptions) {
	for _, catalog := range catalogs {
		en := catalog.en
		if opt.Strip {
			en = stripNL(en)
			en = strings.Join(strings.Fields(en), " ")
		}
		en = strings.Trim(en, "\n")
		if !opt.IsPre && en == "" {
			continue
		}

		if opt.IsPre {
			fmt.Fprintln(w, gchalk.Blue(strings.Trim(catalog.pre, "\n")))
		}
		if !opt.JAOnly {
			fmt.Fprintln(w, gchalk.Green(en))
		}

		if !opt.ENOnly {
			ja := catalog.ja
			if opt.Strip {
				ja = stripNL(ja)
				ja = strings.Join(strings.Fields(ja), " ")
			}
			fmt.Fprintln(w, strings.Trim(ja, "\n"))
		}

		if opt.IsPre {
			fmt.Fprintln(w, gchalk.Blue(strings.Trim(catalog.post, "\n")))
		}
		fmt.Fprintln(w)
	}
}

// 英日辞書の内容をTSV形式で表示する
func TSVList(strip bool, fileNames []string) {
	w := io.Writer(os.Stdout)
	tsvList(w, strip, fileNames)
}

func tsvList(w io.Writer, strip bool, fileNames []string) {
	for _, fileName := range fileNames {
		catalogs, err := loadCatalog(fileName)
		if err != nil {
			log.Println(err)
		}
		writeTSV(w, strip, catalogs)
	}
}

func writeTSV(w io.Writer, strip bool, catalogs []Catalog) {
	for _, catalog := range catalogs {
		en := catalog.en
		if strip {
			en = stripNL(en)
			en = strings.Join(strings.Fields(en), " ")
		} else {
			en = strings.ReplaceAll(en, "\n", " ")
		}
		ja := catalog.ja
		if strip {
			ja = stripNL(ja)
			ja = strings.Join(strings.Fields(ja), " ")
		} else {
			ja = strings.ReplaceAll(ja, "\n", " ")
		}
		fmt.Fprintf(w, "%s\t%s\n", en, ja)
	}
}
