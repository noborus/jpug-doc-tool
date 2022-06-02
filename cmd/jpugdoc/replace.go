package jpugdoc

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/agnivade/levenshtein"
	"github.com/iancoleman/orderedmap"
	"github.com/noborus/go-textra"
)

// type Catalog *orderedmap.OrderedMap

type Rep struct {
	catalog *orderedmap.OrderedMap
	mt      bool
	prompt  bool
	similar int
	api     *textra.TexTra
	apiType string
}

func loadCatalog(fileName string) *orderedmap.OrderedMap {
	src, err := ReadAllFile(fileName)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return nil
	}
	catalog := orderedmap.New()

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

func (rep Rep) Replace(src []byte) []byte {
	ret := rep.paraReplace(src)
	return ret
}

func promptReplace(src []byte, replace []byte) []byte {
	if !prompter.YN("replace?", false) {
		return src
	}
	return REPARA.ReplaceAll(src, []byte(replace))
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

	if ja, ok := rep.catalog.Get(enstr); ok {
		para := fmt.Sprintf("$1<!--\n%s\n-->\n%s$3", en, strings.TrimRight(ja.(string), "\n"))
		if rep.prompt {
			fmt.Println(string(src))
			fmt.Println("前回と一致")
			fmt.Println(ja.(string))
			return promptReplace(src, []byte(para))
		}
		return REPARA.ReplaceAll(src, []byte(para))
	}

	// <literal>.*</literal>又は<literal>.*</literal><returnvalue>.*</returnvalue>又は+programlisting のみのparaだった場合は無視する
	if RELITERAL.Match(src) || RELIRET.Match(src) || RELIRETPROG.Match(src) || RECOMMENTSTART.Match(src) {
		return src
	}
	// 日本語が含まれていた場合は無視
	if REJASTRING.Match(src) {
		return src
	}

	if rep.mt {
		fmt.Printf("API...[%.30s] ", enstr)
		ja, err := rep.api.Translate(rep.apiType, enstr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "replace: %s", err)
			return src
		}
		if ja == "" {
			return src
		}
		ja = KUTEN.ReplaceAllString(ja, "。\n")
		para := fmt.Sprintf("$1<!--\n%s\n-->\n<!-- 《機械翻訳》 -->\n%s$3", en, strings.TrimRight(ja, "\n"))
		fmt.Print("Done\n")

		if rep.prompt {
			fmt.Println(string(src))
			fmt.Println("機械翻訳")
			fmt.Println(string(ja))
			return promptReplace(src, []byte(para))
		}
		return REPARA.ReplaceAll(src, []byte(para))
	}

	if rep.similar == 0 {
		return src
	}

	var maxdis float64
	den := ""
	dja := ""
	keys := rep.catalog.Keys()
	for _, dicen := range keys {
		dicja, _ := rep.catalog.Get(dicen)
		distance := levenshtein.ComputeDistance(enstr, dicen)
		dis := (1 - (float64(distance) / float64(len(enstr)))) * 100
		if dis > maxdis {
			den = dicen
			dja = dicja.(string)
			maxdis = dis
		}
	}
	if maxdis > float64(rep.similar) {
		para := fmt.Sprintf("$1<!--\n%s\n-->\n<!-- マッチ度[%f]\n%s\n-->\n%s$3", en, maxdis, strings.TrimRight(den, "\n"), strings.TrimRight(dja, "\n"))
		if rep.prompt {
			fmt.Println(string(src))
			fmt.Println("類似置き換え")
			fmt.Println(string(dja))
			return promptReplace(src, []byte(para))
		}
		return REPARA.ReplaceAll(src, []byte(para))
	}
	return src
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

func Replace(fileNames []string, mt bool, similar int, prompt bool) {
	apiConfig := textra.Config{}
	apiConfig.ClientID = Config.ClientID
	apiConfig.ClientSecret = Config.ClientSecret
	apiConfig.Name = Config.Name
	cli, err := textra.New(apiConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "textra: %s", err)
	}

	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)

		rep := Rep{
			similar: similar,
			catalog: catalog,
			mt:      mt,
			apiType: Config.APIAutoTranslateType,
		}
		if mt && cli != nil {
			rep.api = cli
		} else {
			rep.mt = false
		}
		rep.prompt = prompt

		src, err := ReadAllFile(fileName)
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
