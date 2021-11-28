package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/spf13/cobra"
)

type Catalog map[string]string

type Rep struct {
	catalog Catalog
	mt      bool
	similar int
	api     MTClient
}

func loadCatalog(fileName string) Catalog {
	src, err := ReadFile(fileName)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return nil
	}
	catalog := make(map[string]string)

	catas := SPLITCATALOG.FindAll(src, -1)
	for _, cata := range catas {
		re := SPLITCATALOG.FindSubmatch(cata)
		en := string(re[1])
		ja := string(re[2])
		catalog[en] = ja
	}
	return catalog
}

func (rep Rep) Replace(src []byte) []byte {
	ret := rep.paraReplace(src)
	return ret
}

func (rep Rep) paraReplace(src []byte) []byte {
	if RECOMMENT.Match(src) {
		return src
	}
	re := REPARA.FindSubmatch(src)

	en := strings.TrimRight(string(re[2]), "\n")

	enstr := strings.ReplaceAll(en, "\n", " ")
	enstr = MultiSpace.ReplaceAllString(enstr, " ")
	enstr = strings.TrimSpace(enstr)

	if ja, ok := rep.catalog[enstr]; ok {
		para := fmt.Sprintf("$1<!--\n%s\n-->\n%s$3", en, strings.TrimRight(ja, "\n"))
		ret := REPARA.ReplaceAll(src, []byte(para))
		return ret
	}

	if rep.mt {
		fmt.Print("API...")
		ja := rep.api.textraTranslate(enstr)
		ja = KUTEN.ReplaceAllString(ja, "。\n")
		para := fmt.Sprintf("$1<!--\n%s\n-->\n<!-- 《機械翻訳》 -->\n%s$3", en, strings.TrimRight(ja, "\n"))
		ret := REPARA.ReplaceAll(src, []byte(para))
		fmt.Print("Done\n")
		return ret
	}

	if rep.similar == 0 {
		return src
	}

	var maxdis float64
	den := ""
	dja := ""
	for dicen, dicja := range rep.catalog {
		distance := levenshtein.ComputeDistance(enstr, dicen)
		dis := (1 - (float64(distance) / float64(len(enstr)))) * 100
		if dis > maxdis {
			den = dicen
			dja = dicja
			maxdis = dis
		}
	}
	if maxdis > float64(rep.similar) {
		para := fmt.Sprintf("$1<!--\n%s\n-->\n<!-- マッチ度[%f]\n%s\n-->\n%s$3", en, maxdis, strings.TrimRight(den, "\n"), strings.TrimRight(dja, "\n"))
		ret := REPARA.ReplaceAll(src, []byte(para))
		return ret
	}
	return src
}

func replace(fileNames []string, mt bool, similar int) {
	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)

		rep := Rep{
			similar: similar,
			catalog: catalog,
			mt:      mt,
		}
		rep.api = apiClient(Config)

		src, err := ReadFile(fileName)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			continue
		}

		ret := REPARA.ReplaceAllFunc(src, rep.Replace)
		if bytes.Equal(src, ret) {
			continue
		}

		if err := rewriteFile(fileName, ret); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
	}
}

//  file rewrite.
func rewriteFile(fileName string, body []byte) error {
	fmt.Printf("replace: %s\n", fileName)
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	fmt.Fprint(out, string(body))
	out.Close()
	return nil
}

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "英語のパラグラフを「<!--英語-->日本語翻訳」に置き換える",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var mt bool
		var similar int
		var err error
		if similar, err = cmd.PersistentFlags().GetInt("similar"); err != nil {
			log.Println(err)
			return
		}
		if mt, err = cmd.PersistentFlags().GetBool("mt"); err != nil {
			log.Println(err)
			return
		}
		if len(args) > 0 {
			replace(args, mt, similar)
			return
		}

		fileNames := targetFileName()
		replace(fileNames, mt, similar)
	},
}

func init() {
	rootCmd.AddCommand(replaceCmd)
	replaceCmd.PersistentFlags().IntP("similar", "s", 0, "Degree of similarity")
	replaceCmd.PersistentFlags().BoolP("mt", "", false, "Use machine translation")
}
