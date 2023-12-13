package jpugdoc

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var Version = "dev"

var DicDir = "./.jpug-doc-tool/"

type apiConfig struct {
	ClientID             string
	ClientSecret         string
	Name                 string
	APIAutoTranslate     string
	APIAutoTranslateType string
}

var Config apiConfig

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
	"libpq0.sgml":    {},
	"libpq1.sgml":    {},
	"libpq2.sgml":    {},
	"libpq3.sgml":    {},
}

func IgnoreFileNames(fileNames []string) []string {
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

	if _, err := os.Stat(DicDir); os.IsNotExist(err) {
		err := os.Mkdir(DicDir, 0o755)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	refdir := DicDir + "/ref"
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

func loadCatalog(fileName string) []Catalog {
	catalog := make([]Catalog, 0)

	src, err := ReadAllFile(fileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return catalog
	}

	catas := SPLITCATALOG.FindAll(src, -1)
	for _, cata := range catas {
		re := SPLITCATALOG.FindSubmatch(cata)
		c := Catalog{
			pre:      string(re[1]),
			en:       string(re[2]),
			ja:       string(re[3]),
			cdatapre: string(re[4]),
		}
		catalog = append(catalog, c)
	}
	return catalog
}

func getDiff(vTag string, fileName string) []byte {
	args := []string{"diff", "--histogram", "-U10000", vTag, fileName}
	cmd := exec.Command("git", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("exec", err)
	}

	var src []byte
	cmd.Start()
	src, err = io.ReadAll(stdout)
	if err != nil {
		log.Fatal("read", err)
	}
	cmd.Wait()
	return src
}
