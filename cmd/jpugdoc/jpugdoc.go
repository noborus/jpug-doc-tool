package jpugdoc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
		os.Mkdir(DICDIR, 0o755)
	}
	refdir := DICDIR + "/ref"
	if _, err := os.Stat(refdir); os.IsNotExist(err) {
		os.Mkdir(refdir, 0o755)
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
