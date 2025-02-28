package jpugdoc

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var Version = "dev"

var DicDir = filepath.Join(".", ".jpug-doc-tool/")

var versionFile = "version.sgml"

var (
	ErrVersionTag = fmt.Errorf("version tag error")
	ErrMemberName = fmt.Errorf("member name error")
	ErrNoMatch    = fmt.Errorf("no match")
)

var Verbose bool

type jpugDocConfig struct {
	APIKEY               string
	APISecret            string
	APIName              string
	APIAutoTranslate     string
	APIAutoTranslateType string
}

var Config jpugDocConfig

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

	refDir := filepath.Join(DicDir, "ref")
	if _, err := os.Stat(refDir); os.IsNotExist(err) {
		err := os.Mkdir(refDir, 0o755)
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

// saveCatalog saves the specified catalog to a file.
func saveCatalog(fileName string, catalogs Catalogs) {
	catalogName := filepath.Join(DicDir, fileName+".t")
	f, err := os.Create(catalogName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, catalog := range catalogs {
		fmt.Fprintf(f, "␝%s␟", catalog.pre)
		fmt.Fprintf(f, "%s␟", catalog.en)
		fmt.Fprintf(f, "%s␞", catalog.ja)
		fmt.Fprintf(f, "%s␞", catalog.preCDATA)
		fmt.Fprintf(f, "%s␞\n", catalog.post)
	}
}

// loadCatalog loads a catalog from the specified file.
func loadCatalog(fileName string) ([]Catalog, error) {
	catalogName := filepath.Join(DicDir, fileName+".t")
	src, err := ReadAllFile(catalogName)
	if err != nil {
		return nil, err
	}

	return getCatalogs(src), nil
}

func getCatalogs(src []byte) []Catalog {
	catalogs := make([]Catalog, 0)
	catas := SPLITCATALOG.FindAll(src, -1)
	for _, cata := range catas {
		re := SPLITCATALOG.FindSubmatch(cata)
		c := Catalog{
			pre:      string(re[1]),
			en:       string(re[2]),
			ja:       string(re[3]),
			preCDATA: string(re[4]),
			post:     string(re[5]),
		}
		catalogs = append(catalogs, c)
	}
	return catalogs
}

// version.sgmlからバージョンタグを取得する
// 15.4 → REL_15_4
func versionTag() (string, error) {
	cmd := exec.Command("make", "version.sgml")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %v", ErrVersionTag, err)
	}

	src, err := ReadAllFile(versionFile)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrVersionTag, err)
	}
	return version(src)
}

func version(src []byte) (string, error) {
	ver := regexp.MustCompile(`<!ENTITY version "([0-9][0-9]+)([^\"]*)">`)
	re := ver.FindSubmatch(src)
	if len(re) < 1 {
		return "master", nil
	}
	v := strings.ReplaceAll(strings.ToUpper(string(re[2])), ".", "_")
	if strings.Contains(v, "DEVEL") {
		return "master", nil
	}
	tag := fmt.Sprintf("REL_%s_%s", string(re[1]), strings.TrimLeft(v, "_"))
	return tag, nil
}

func getMemberName() (map[string]bool, error) {
	m, err := filepath.Glob("release-*.sgml")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMemberName, err)
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("%w: release-*.sgmlが見つかりません", ErrMemberName)
	}
	f, err := os.Open(m[len(m)-1])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMemberName, err)
	}
	return getMember(f), nil
}

func getMember(f *os.File) map[string]bool {
	members := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<member>") {
			m := strings.TrimSpace(line)
			m = strings.TrimPrefix(m, "<member>")
			m = strings.TrimSuffix(m, "</member>")
			members[m] = true
		}
	}
	return members
}
