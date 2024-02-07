package jpugdoc

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/agnivade/levenshtein"
	"github.com/noborus/go-textra"
)

type Rep struct {
	catalog []Catalog
	update  bool
	mt      bool
	prompt  bool
	similar int
	api     *textra.TexTra
	apiType string
}

// Replace は指定されたファイル名のファイルを置き換える
func Replace(fileNames []string, vTag string, update bool, mt bool, similar int, prompt bool) {
	apiConfig := textra.Config{}
	apiConfig.ClientID = Config.ClientID
	apiConfig.ClientSecret = Config.ClientSecret
	apiConfig.Name = Config.Name
	cli, err := textra.New(apiConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "textra: %s", err)
	}
	if update && vTag == "" {
		v, err := versionTag()
		if err != nil {
			log.Fatal(err)
		}
		vTag = v
	}

	for _, fileName := range fileNames {
		catalog := loadCatalog(fileName)

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

		var ret []byte
		if rep.update {
			ret = rep.updateFromCatalog(fileName, vTag, src)
		} else {
			ret = rep.replaceCatalogs(src)
			// <para>のみ更に置き換える
			ret = REPARA.ReplaceAllFunc(ret, rep.paraReplace)
		}

		if bytes.Equal(src, ret) {
			continue
		}

		if err := rewriteFile(fileName, ret); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
	}
}

// Replace all from catalog
func (rep Rep) replaceCatalogs(src []byte) []byte {
	for _, c := range rep.catalog {
		if c.en != "" {
			src = rep.replaceCatalog(src, c)
		}
	}
	// indexterm
	for _, c := range rep.catalog {
		if c.en == "" {
			src = rep.additionalReplace(src, c)
		}
	}
	return src
}

// .jpug-doc-tool/filename.sgml.t のカタログを使用して置き換える
func (rep Rep) replaceCatalog(src []byte, c Catalog) []byte {
	cen := append([]byte(c.en), '\n')
	hen := REVHIGHHUN2.ReplaceAll(cen, []byte("&#45;&#45;-"))
	hen = REVHIGHHUN.ReplaceAll(hen, []byte("&#45;-"))
	p := 0
	pp := 0
	ret := make([]byte, 0)
	for p < len(src) {
		pp = bytes.Index(src[p:], cen)
		if pp == -1 {
			ret = append(ret, src[p:]...)
			break
		}

		if !inComment(src[:p+pp]) {
			ret = append(ret, src[p:p+pp]...)
			if inCDATA(src[:p+pp]) {
				ret = append(ret, []byte(c.cdatapre+"]]><!--\n")...)
			} else {
				ret = append(ret, []byte("<!--\n")...)
			}
			ret = append(ret, hen...)
			if inCDATA(src[:p+pp]) {
				ret = append(ret, []byte("--><![CDATA[\n")...)
			} else {
				ret = append(ret, []byte("-->\n")...)
			}

			if c.ja != "" {
				ret = append(ret, []byte(c.ja)...)
				ret = append(ret, src[p+pp+len(c.en):]...)
			} else {
				ret = append(ret, src[p+pp+len(c.en)+1:]...)
			}

			break
		}
		// Already in Japanese.
		ret = append(ret, src[p:p+pp+len(c.en)]...)
		p = p + pp + len(c.en)
	}
	return ret
}

func inComment(src []byte) bool {
	s := bytes.LastIndex(src, []byte("<!--"))
	e := bytes.LastIndex(src, []byte("-->"))
	return s > e
}

func inCDATA(src []byte) bool {
	s := bytes.LastIndex(src, []byte("<![CDATA["))
	e := bytes.LastIndex(src, []byte("]]>"))
	return s > e
}

// コメント後の翻訳文の形式以外の追加文。
func (rep Rep) additionalReplace(src []byte, c Catalog) []byte {
	p := foundReplace(src, c)
	if p == -1 {
		return src
	}
	ret := make([]byte, 0)
	ret = append(ret, src[:p+len(c.pre)]...)
	ret = append(ret, c.ja...)
	ret = append(ret, '\n')
	ret = append(ret, src[p+len(c.pre):]...)
	return ret
}

func foundReplace(src []byte, c Catalog) int {
	for p := 0; p < len(src); {
		i := bytes.Index(src[p:], []byte(c.pre))
		if i == -1 {
			return -1
		}
		j := bytes.Index(src[p+i:], []byte("\n"+c.ja))
		if j == -1 {
			// before conversion.
			return p + i
		}
		// Already converted.
		p = p + i + j + len(c.pre) + 1
	}
	return -1
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
	pair, left, err := enjaPair(src)
	if err != nil {
		return src
	}
	en, _, _ := splitComment(src)
	for _, c := range rep.catalog {
		if stripEN(c.en) == pair.en {
			if c.ja == pair.ja {
				return src
			}
			fmt.Println("更新", pair.en)
			para := fmt.Sprintf("$1<!--%s-->\n%s$3", en, strings.TrimRight(c.ja, "\n"))
			ret := REPARA.ReplaceAll(src, []byte(para))
			fmt.Println(string(left))
			ret = append(ret, left...)
			return ret
		}
	}
	return src
}

func replaceCatalog(src []byte, catalogs []Catalog, o Catalog) []byte {
	var ja string
	for _, c := range catalogs {
		if o.en != "" && c.en != "" && o.ja != "" && c.en == o.en {
			if c.ja == o.ja {
				return src
			}
			ja = c.ja
		}
	}
	if ja != "" {
		return bytes.ReplaceAll(src, []byte(o.ja), []byte(ja))
	}
	return src
}

// srcを新しいcatalogの日本語訳に置き換えて更新する
func (rep Rep) updateFromCatalog(fileName string, vTag string, src []byte) []byte {
	srcDiff := getDiff(vTag, fileName)
	org := Extraction(srcDiff)
	for _, o := range org {
		src = replaceCatalog(src, rep.catalog, o)
	}
	return src
}

// <para>の置き換え
func (rep Rep) paraReplace(src []byte) []byte {
	re := REPARA.FindSubmatch(src)
	para := string(re[1])
	en := strings.TrimRight(string(re[2]), "\n")
	en = REVHIGHHUN2.ReplaceAllString(en, "&#45;&#45;-")
	en = REVHIGHHUN.ReplaceAllString(en, "&#45;-")
	enStr := stripEN(string(re[2]))
	/*
		for _, c := range rep.catalog {
			if c.en == "" {
				continue
			}
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
	*/
	if strings.HasPrefix(enStr, "<!--") {
		return src
	}
	if strings.Contains(enStr, "<para>") && strings.Contains(enStr, "<!--") {
		return src
	}
	if NIHONGO.MatchString(enStr) {
		return src
	}
	// <para>\nで改行されていない場合はスキップ
	if !strings.Contains(para, "\n") {
		return src
	}
	if bytes.Contains(src, []byte("<returnvalue>")) {
		return src
	}
	log.Println("replace?", enStr)
	if rep.similar > 0 && rep.mt {
		log.Println("simMtReplace", en)
		return rep.simMtReplace(src, en, enStr)
	}
	if rep.similar != 0 {
		return rep.simReplace(src, en, enStr)
	}
	if rep.mt {
		return rep.mtReplace(src, en, enStr)
	}
	return src
}

// 機械翻訳による置き換え
func (rep Rep) mtReplace(src []byte, en string, enStr string) []byte {
	ja := rep.MTtrans(enStr)
	if ja == "" {
		return src
	}
	para := fmt.Sprintf("$1<!--\n%s\n-->\n《機械翻訳》%s$3", en, strings.TrimRight(ja, "\n"))
	fmt.Print("Done\n")

	if !rep.prompt {
		return REPARA.ReplaceAll(src, []byte(para))
	}

	fmt.Println(string(src))
	fmt.Println("機械翻訳")
	fmt.Println(string(ja))
	return promptReplace(src, []byte(para))
}

// 機械翻訳
func (rep Rep) MTtrans(enStr string) string {
	fmt.Printf("API...[%.30s] ", enStr)
	ja, err := rep.api.Translate(rep.apiType, enStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replace: %s\n", err)
		return ""
	}
	if ja == "" {
		return ""
	}
	ja = KUTEN.ReplaceAllString(ja, "。\n")
	return ja
}

// 類似文置き換え
func (rep Rep) simReplace(src []byte, en string, enstr string) []byte {
	var maxdis float64
	//den := ""
	dja := ""
	for _, c := range rep.catalog {
		distance := levenshtein.ComputeDistance(enstr, c.en)
		dis := (1 - (float64(distance) / float64(len(enstr)))) * 100
		if dis > maxdis {
			//den = c.en
			dja = c.ja
			maxdis = dis
		}
	}
	if maxdis > float64(rep.similar) {
		para := fmt.Sprintf("$1<!--\n%s\n-->\n《マッチ度[%f]》%s$3", en, maxdis, strings.TrimRight(dja, "\n"))
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

// 類似文置き換え
func (rep Rep) simMtReplace(src []byte, en string, enstr string) []byte {
	var maxdis float64
	//den := ""
	dja := ""
	for _, c := range rep.catalog {
		distance := levenshtein.ComputeDistance(enstr, c.en)
		dis := (1 - (float64(distance) / float64(len(enstr)))) * 100
		if dis > maxdis {
			//den = c.en
			dja = c.ja
			maxdis = dis
		}
	}
	if maxdis > float64(rep.similar) {
		simja := strings.TrimRight(dja, "\n")
		mtja := ""
		para := ""
		if maxdis < 90 {
			mtja = rep.MTtrans(enstr)
		}
		if mtja != "" {
			mtja = strings.TrimRight(mtja, "\n")
			para = fmt.Sprintf("$1<!--\n%s\n-->\n《マッチ度[%f]》%s\n《機械翻訳》%s$3", en, maxdis, simja, mtja)
		} else {
			para = fmt.Sprintf("$1<!--\n%s\n-->\n《マッチ度[%f]》%s$3", en, maxdis, simja)
		}
		return REPARA.ReplaceAll(src, []byte(para))
	}

	mtja := rep.MTtrans(enstr)
	if mtja != "" {
		mtja = strings.TrimRight(mtja, "\n")
		para := fmt.Sprintf("$1<!--\n%s\n-->\n《機械翻訳》%s$3", en, mtja)
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
