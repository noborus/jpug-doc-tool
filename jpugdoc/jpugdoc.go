package jpugdoc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/elliotchance/orderedmap/v2"
)

var Version = "dev"

var DICDIR = "./.jpug-doc-tool/"

type apiConfig struct {
	ClientID             string
	ClientSecret         string
	Name                 string
	APIAutoTranslate     string
	APIAutoTranslateType string
}

var Config apiConfig

func IgnoreFileNames(fileNames []string) []string {
	var ignoreFile map[string]struct{} = map[string]struct{}{
		"jpug-doc.sgml":  {},
		"config0.sgml":   {},
		"config1.sgml":   {},
		"config2.sgml":   {},
		"config3.sgml":   {},
		"func0.sgml":     {},
		"func1.sgml":     {},
		"func2.sgml":     {},
		"func3.sgml":     {},
		"func4.sgml":     {},
		"catalogs0.sgml": {},
		"catalogs1.sgml": {},
		"catalogs2.sgml": {},
		"catalogs3.sgml": {},
		"catalogs4.sgml": {},
	}

	ret := make([]string, 0, len(fileNames))
	for _, fileName := range fileNames {
		if _, ok := ignoreFile[fileName]; ok {
			continue
		}
		ret = append(ret, fileName)
	}
	return ret
}

func InitJpug() {
	f, err := filepath.Glob("./*.sgml")
	if err != nil || len(f) == 0 {
		fmt.Fprintln(os.Stderr, "*sgmlファイルがあるディレクトリで実行してください")
		fmt.Fprintln(os.Stderr, "cd github.com/pgsql-jp/jpug-doc/doc/src/sgml")
		return
	}
	if _, err := os.Stat(DICDIR); os.IsNotExist(err) {
		err := os.Mkdir(DICDIR, 0o755)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	refdir := DICDIR + "/ref"
	if _, err := os.Stat(refdir); os.IsNotExist(err) {
		err := os.Mkdir(refdir, 0o755)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func ReadAllFile(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return src, nil
}

func loadCatalog(fileName string) *orderedmap.OrderedMap[string, string] {
	catalog := orderedmap.NewOrderedMap[string, string]()
	src, err := ReadAllFile(fileName)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return catalog
	}

	catas := SPLITCATALOG.FindAll(src, -1)
	for _, cata := range catas {
		re := SPLITCATALOG.FindSubmatch(cata)
		if len(re[1]) == 0 {
			catalog.Set(string(re[2]), "")
			continue
		}
		en := string(re[1])
		ja := string(re[2])
		catalog.Set(en, ja)
	}
	return catalog
}
