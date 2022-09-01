package jpugdoc

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/agnivade/levenshtein"
	"github.com/noborus/go-textra"
)

// type Catalog *orderedmap.OrderedMap[string, string]

type Rep struct {
	catalog []Catalog
	update  bool
	mt      bool
	prompt  bool
	similar int
	api     *textra.TexTra
	apiType string
}

func (rep Rep) Replace(src []byte) []byte {
	if rep.mt || rep.similar > 0 {
		src = rep.paraReplace(src)
	}

	if rep.update {
		src = rep.updateReplace(src)
	}
	return src
}

func promptReplace(src []byte, replace []byte) []byte {
	if !prompter.YN("replace?", false) {
		return src
	}
	return REPARA.ReplaceAll(src, []byte(replace))
}

func stripEN(src string) string {
	str := strings.TrimRight(src, "\n")
	str = strings.ReplaceAll(str, "\n", " ")
	str = MultiSpace.ReplaceAllString(str, " ")
	str = strings.TrimSpace(str)
	return str
}

func (rep Rep) updateReplace(src []byte) []byte {
	pair, _, err := enjaPair(src)
	if err != nil {
		return src
	}
	en, _, _ := splitComment(src)
	enstr := stripEN(string(en))
	for _, c := range rep.catalog {
		if stripEN(c.en) == pair.en {
			if c.ja == pair.ja {
				return src
			}
			fmt.Println("更新", pair.en)
			para := fmt.Sprintf("$1<!--\n%s\n-->\n%s$3", enstr, strings.TrimRight(c.ja, "\n"))
			return REPARA.ReplaceAll(src, []byte(para))
		}
	}

	return src
}

func (rep Rep) replaceAll(src []byte, c Catalog) []byte {
	cen := append([]byte(c.en), '\n')
	en := REVHIGHHUN.ReplaceAllString(c.en, "&#45;-")
	p := 0
	pp := 0
	ret := make([]byte, 0)
	for p < len(src) {
		pp = bytes.Index(src[p:], cen)
		if pp == -1 {
			ret = append(ret, src[p:]...)
			break
		}

		if string(src[(p+pp)-5:(p+pp)]) == "<!--\n" {
			ret = append(ret, src[p:p+pp+len(c.en)]...)
			p = p + pp + len(c.en)
			continue
		}

		ret = append(ret, src[p:p+pp]...)
		ret = append(ret, []byte("<!--\n")...)
		ret = append(ret, []byte(en)...)
		ret = append(ret, '\n')
		ret = append(ret, []byte("-->\n")...)
		ret = append(ret, []byte(c.ja)...)

		p = p + pp + len(c.en)
		/*
			m := bytes.Index(src[p:], []byte(c.en))
			if m == -1 {
				ret = append(ret, src[p:]...)
				//fmt.Println(string(ret))
				break
			}
			p = p + m + 1
			fmt.Println(p)
			ret = append(ret, src[:p]...)
			ret = append(ret, []byte("<!--\n")...)
			ret = append(ret, []byte(en)...)
			ret = append(ret, '\n')
			ret = append(ret, []byte("-->\n")...)
			ret = append(ret, []byte(c.ja)...)
		*/
	}
	//fmt.Println(string(ret))
	return ret
}

func (rep Rep) replaceCatalog(src []byte) []byte {
	for _, c := range rep.catalog {
		src = rep.replaceAll(src, c)
	}
	return src
}

func (rep Rep) paraReplace(src []byte) []byte {
	re := REPARA.FindSubmatch(src)
	en := strings.TrimRight(string(re[2]), "\n")
	enstr := stripEN(string(re[2]))
	for _, c := range rep.catalog {
		if stripEN(c.en) == enstr {
			para := fmt.Sprintf("$1<!--\n%s\n-->\n%s$3", en, strings.TrimRight(c.ja, "\n"))
			if rep.prompt {
				fmt.Println(string(src))
				fmt.Println("前回と一致")
				fmt.Println(c.ja)
				return promptReplace(src, []byte(para))
			}
			return REPARA.ReplaceAll(src, []byte(para))
		}
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
	for _, c := range rep.catalog {

		distance := levenshtein.ComputeDistance(enstr, c.en)
		dis := (1 - (float64(distance) / float64(len(enstr)))) * 100
		if dis > maxdis {
			den = c.en
			dja = c.ja
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

// file rewrite.
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

func Replace(fileNames []string, update bool, mt bool, similar int, prompt bool) {
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
			update:  update,
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

		ret := rep.replaceCatalog(src)
		/*
			ret := REPARA.ReplaceAllFunc(src, rep.Replace)
			if bytes.Equal(src, ret) {
				continue
			}
		*/
		if err := rewriteFile(fileName, ret); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
	}
}
