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

// Rep は置き換えを行う構造体
type Rep struct {
	catalogs []Catalog
	vTag     string
	update   bool
	similar  int
	mt       int
	prompt   bool
	api      *textra.TexTra
	apiType  string
	// 機械翻訳エラー
	err error
}

func newOpts(vTag string, update bool, similar int, mt int, prompt bool) (*Rep, error) {
	rep := &Rep{
		similar: similar,
		update:  update,
		mt:      mt,
		apiType: Config.APIAutoTranslateType,
	}
	if update && vTag == "" {
		v, err := versionTag()
		if err != nil {
			return rep, err
		}
		rep.vTag = v
	}

	mtCli, _ := newTextra(Config)
	if mt > 0 && mtCli != nil {
		rep.mt = mt
		rep.api = mtCli
	} else {
		rep.mt = 0
		rep.api = nil
	}

	rep.prompt = prompt
	return rep, nil
}

// Replace は指定されたファイル名のファイルを置き換える
func Replace(fileNames []string, vTag string, update bool, similar int, mt int, prompt bool) error {
	rep, err := newOpts(vTag, update, similar, mt, prompt)
	if err != nil {
		return err
	}
	if Verbose {
		log.Printf("similar %d mt %d\n", rep.similar, rep.mt)
	}

	for _, fileName := range fileNames {
		rep.catalogs, err = loadCatalog(fileName)
		if err != nil {
			log.Print(err.Error())
			continue
		}
		ret, err := rep.replace(fileName)
		if err != nil {
			// 機械翻訳エラーだった場合は終了
			if rep.err != nil {
				log.Fatal(rep.err)
			}
			log.Print(err.Error())
			continue
		}
		if ret == nil {
			continue
		}

		if err := rewriteFile(fileName, ret); err != nil {
			return fmt.Errorf("rewrite: %s: %w", fileName, err)
		}
		fmt.Printf("replace: %s\n", fileName)
	}
	return nil
}

// replace はファイルを置き換えた結果を返す
func (rep *Rep) replace(fileName string) ([]byte, error) {
	src, err := ReadAllFile(fileName)
	if err != nil {
		log.Printf("skip:%s:%s\n", fileName, err)
		return nil, nil
	}

	ret, err := rep.replaceAll(fileName, src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fileName, err)
	}
	// 置き換えがない場合はnilを返す
	if bytes.Equal(src, ret) {
		return nil, nil
	}
	return ret, nil
}

func (rep *Rep) replaceAll(fileName string, src []byte) ([]byte, error) {
	// 更新の場合はすでに翻訳がある箇所を更新する
	if rep.update {
		return rep.updateFromCatalog(fileName, src)
	}

	// 一致文置き換え
	ret := rep.replaceCatalogs(src)

	// <para>のみ更に置き換える
	ret = REPARA.ReplaceAllFunc(ret, rep.paraReplace)
	if rep.err != nil {
		return nil, fmt.Errorf("%s: %w", fileName, rep.err)
	}
	return ret, nil
}

func newTextra(apiConfig apiConfig) (*textra.TexTra, error) {
	if apiConfig.ClientID == "" || apiConfig.ClientSecret == "" {
		return nil, fmt.Errorf("textra: ClientID and ClientSecret are required")
	}
	config := textra.Config{}
	config.ClientID = apiConfig.ClientID
	config.ClientSecret = apiConfig.ClientSecret
	config.Name = apiConfig.Name
	return textra.New(config)
}

// Replace all from catalog
func (rep *Rep) replaceCatalogs(src []byte) []byte {
	// 追加形式の翻訳文を追加
	for _, catalog := range rep.catalogs {
		if catalog.en == "" {
			src = rep.additionalReplace(src, catalog)
		}
	}
	// コメント形式の翻訳文を追加
	for _, catalog := range rep.catalogs {
		if catalog.en != "" {
			src = rep.replaceCatalog(src, catalog)
		}
	}
	return src
}

// カタログを一つづつ置き換える
func (rep Rep) replaceCatalog(src []byte, catalog Catalog) []byte {
	if catalog.ja == "no translation" {
		return src
	}
	cen := append([]byte(catalog.en), '\n')
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

		if catalog.post != "" {
			start := p + pp + len(cen)
			if start > len(src) {
				ret = append(ret, src[p:]...)
				break
			}
			if !bytes.HasPrefix(src[start:], []byte(catalog.post)) {
				ret = append(ret, src[p:p+pp+len(catalog.en)]...)
				p = p + pp + len(catalog.en)
				continue
			}
		}

		if !inComment(src[:p+pp]) {
			ret = append(ret, src[p:p+pp]...)
			if inCDATA(src[:p+pp]) {
				ret = append(ret, []byte(catalog.preCDATA+"]]><!--\n")...)
			} else {
				ret = append(ret, []byte("<!--\n")...)
			}
			ret = append(ret, hen...)
			if inCDATA(src[:p+pp]) {
				ret = append(ret, []byte("--><![CDATA[\n")...)
			} else {
				ret = append(ret, []byte("-->\n")...)
			}

			if catalog.ja != "" {
				ret = append(ret, []byte(catalog.ja)...)
				ret = append(ret, src[p+pp+len(catalog.en):]...)
			} else {
				ret = append(ret, src[p+pp+len(catalog.en)+1:]...)
			}
			break
		}

		// Already in Japanese.
		ret = append(ret, src[p:p+pp+len(catalog.en)]...)
		p = p + pp + len(catalog.en)
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
func (rep Rep) additionalReplace(src []byte, catalog Catalog) []byte {
	p := foundReplace(src, catalog)
	if p == -1 {
		return src
	}
	ret := make([]byte, 0)
	ret = append(ret, src[:p+len(catalog.pre)]...)
	ret = append(ret, catalog.ja...)
	ret = append(ret, '\n')
	ret = append(ret, src[p+len(catalog.pre):]...)
	return ret
}

func foundReplace(src []byte, catalog Catalog) int {
	for p := 0; p < len(src); {
		i := bytes.Index(src[p:], []byte(catalog.pre))
		if i == -1 {
			return -1
		}
		j := bytes.Index(src[p+i:], []byte("\n"+catalog.ja))
		if j == -1 {
			// before conversion.
			return p + i
		}
		// Already converted.
		p = p + i + j + len(catalog.pre) + 1
	}
	return -1
}

// promptReplace は置き換えを質問する
func promptReplace(src []byte, replace []byte) []byte {
	if !prompter.YN("replace?", false) {
		return src
	}
	return REPARA.ReplaceAll(src, []byte(replace))
}

// stripNL は改行を削除して空白を一つにする
func stripNL(src string) string {
	str := strings.TrimRight(src, "\n")
	str = strings.ReplaceAll(str, "\n", " ")
	str = MultiSpace.ReplaceAllString(str, " ")
	str = strings.TrimSpace(str)
	return str
}

func (rep Rep) updateReplace(src []byte) []byte {
	pair, left, err := newCatalog(src)
	if err != nil {
		return src
	}
	en, _, _ := splitComment(src)
	for _, c := range rep.catalogs {
		if stripNL(c.en) == pair.en {
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

func updateReplaceCatalog(src []byte, catalogs []Catalog, o Catalog) []byte {
	var ja string
	for _, c := range catalogs {
		if c.en == "" || c.ja == "" || o.en == "" || o.ja == "" {
			continue
		}
		if c.en == o.en {
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
func (rep Rep) updateFromCatalog(fileName string, src []byte) ([]byte, error) {
	srcDiff, err := getDiff(rep.vTag, fileName)
	if err != nil {
		return nil, err
	}
	org := Extraction(srcDiff)
	for _, o := range org {
		src = updateReplaceCatalog(src, rep.catalogs, o)
	}
	return src, nil
}

// <para>の置き換え
func (rep *Rep) paraReplace(src []byte) []byte {
	re := REPARA.FindSubmatch(src)
	para := string(re[1])
	en := strings.TrimRight(string(re[2]), "\n")
	en = REVHIGHHUN2.ReplaceAllString(en, "&#45;&#45;-")
	en = REVHIGHHUN.ReplaceAllString(en, "&#45;-")
	enStr := stripNL(string(re[2]))

	// 既に翻訳済みの場合はスキップ
	if strings.HasPrefix(enStr, "<!--") {
		return src
	}
	if strings.Contains(enStr, "<para>") && strings.Contains(enStr, "<!--") {
		return src
	}
	// 空白のみの場合はスキップ
	if strings.Trim(enStr, " ") == "" {
		return src
	}

	// 既に日本語がある場合はスキップ
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

	// 類似文、機械翻訳置き換え
	ret, err := rep.simMtReplace(src, en, enStr)
	if err != nil {
		log.Println(err.Error())
		rep.err = err
	}
	return ret
}

// 機械翻訳
func (rep *Rep) MTtrans(enStr string) (string, error) {
	if Verbose {
		fmt.Printf("API...[%.30s] ", enStr)
	}
	ja, err := rep.api.Translate(rep.apiType, enStr)
	if err != nil {
		rep.api = nil
		return "", fmt.Errorf("replace: %w", err)
	}
	if ja == "" {
		return "", fmt.Errorf("replace: No translation")
	}
	if Verbose {
		fmt.Printf("Done\n")
	}

	ja = KUTEN.ReplaceAllString(ja, "。\n")
	return ja, nil
}

// 類似文、機械翻訳置き換え
func (rep *Rep) simMtReplace(src []byte, en string, enStr string) ([]byte, error) {
	var maxScore float64
	var simJa string

	// 最も類似度の高い日本語を探す
	for _, c := range rep.catalogs {
		distance := levenshtein.ComputeDistance(enStr, c.en)
		score := (1 - (float64(distance) / float64(len(enStr)))) * 100
		if score > maxScore {
			simJa = c.ja
			maxScore = score
		}
	}
	if simJa == "no translation" {
		return src, nil
	}
	if maxScore > float64(rep.similar) {
		if Verbose {
			fmt.Printf("Similar...[%f][%.30s]\n", maxScore, enStr)
		}
		simJa = strings.TrimRight(simJa, "\n")
	} else {
		simJa = ""
	}

	mtJa := ""
	para := ""
	if maxScore < float64(rep.mt) {
		ja, err := rep.MTtrans(enStr)
		if err != nil {
			return nil, err
		}
		mtJa = ja
	}

	if simJa != "" && mtJa != "" {
		mtJa = strings.TrimRight(mtJa, "\n")
		para = fmt.Sprintf("$1<!--\n%s\n-->\n《マッチ度[%f]》%s\n《機械翻訳》%s$3", en, maxScore, simJa, mtJa)
	} else if simJa != "" {
		para = fmt.Sprintf("$1<!--\n%s\n-->\n《マッチ度[%f]》%s$3", en, maxScore, simJa)
	} else if mtJa != "" {
		para = fmt.Sprintf("$1<!--\n%s\n-->\n《機械翻訳》%s$3", en, mtJa)
	} else {
		return src, nil
	}
	if !rep.prompt {
		return REPARA.ReplaceAll(src, []byte(para)), nil
	}

	return promptReplace(src, []byte(para)), nil
}

// file rewrite.
func rewriteFile(fileName string, body []byte) error {
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = fmt.Fprint(out, string(body))
	return err
}
