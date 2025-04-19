package jpugdoc

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/agnivade/levenshtein"
	"github.com/noborus/go-textra"
)

const (
	MTTransStart = "«"
	MTTransEnd   = "»"
	MTL          = len(MTTransStart)
)

// Rep は置き換えを行う構造体
type Rep struct {
	common   []Catalog
	catalogs []Catalog
	vTag     string
	update   bool
	similar  int
	mt       int
	wip      bool
	prompt   bool
}

// NewRep はRep構造体を初期化する関数
func NewRep(vTag string, update bool, similar int, mt int, prompt bool) (*Rep, error) {
	rep := &Rep{
		similar: similar,
		update:  update,
		mt:      mt,
		prompt:  prompt,
	}
	if update && vTag == "" {
		v, err := versionTag()
		if err != nil {
			return rep, err
		}
		rep.vTag = v
	} else {
		rep.vTag = vTag
	}

	common, err := loadCatalog("common")
	if err != nil {
		return rep, err
	}
	rep.common = regCompile(common)
	return rep, nil
}

// Replace は指定されたファイル名のファイルを置き換える
func Replace(fileNames []string, vTag string, update bool, similar int, mt int, wip bool, prompt bool) error {
	rep, err := NewRep(vTag, update, similar, mt, prompt)
	if err != nil {
		return err
	}
	if Verbose {
		log.Printf("マッチ度 %d 以上を採用。マッチ度 %d 以下であれば機械翻訳を追加\n", rep.similar, rep.mt)
	}

	for _, fileName := range fileNames {
		rep.catalogs, err = loadCatalog(fileName)
		if err != nil {
			log.Print(err.Error())
			// （新規ファイル）カタログがなくても続行
		}
		ret, err := rep.replace(fileName)
		if err != nil {
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
	ret := rep.matchReplace(src)

	if rep.similar == 0 && rep.mt == 0 {
		return ret, nil
	}
	// 類似文、機械翻訳置き換え
	return rep.unmatchedReplace(fileName, ret)
}

// 一致文置き換え
func (rep *Rep) matchReplace(src []byte) []byte {
	// 追加形式の翻訳文を追加
	for _, catalog := range rep.catalogs {
		if catalog.en == "" {
			src = matchAdditional(src, catalog)
		}
	}
	// 共通の翻訳文を追加
	for _, catalog := range rep.common {
		if catalog.enReg != nil {
			src = matchCommon(src, catalog)
		}
	}
	// コメント形式の翻訳文を追加
	for _, catalog := range rep.catalogs {
		if catalog.en != "" {
			src = matchComment(src, catalog)
		}
	}
	return src
}

// カタログを一つずつ置き換える
func matchComment(src []byte, catalog Catalog) []byte {
	if catalog.ja == "no translation" {
		return src
	}
	cen := append([]byte(catalog.en), '\n')
	hen := REVHIGHHUN2.ReplaceAll(cen, []byte("&#45;&#45;-"))
	hen = REVHIGHHUN.ReplaceAll(hen, []byte("&#45;-"))

	p := 0
	pp := 0
	ret := make([]byte, 0, len(src)*2)
	for p < len(src) {
		pp = bytes.Index(src[p:], cen)
		if pp == -1 {
			ret = append(ret, src[p:]...)
			break
		}
		count := 0
		if catalog.ja != "" && catalog.ja[0] == ' ' {
			count = countLeadingSpaces(src, p+pp)
		}
		pp -= count
		if catalog.post == "" {
			// 一致前が改行でない場合はスキップ
			if p+pp < len(src) && p+pp > 0 {
				if src[p+pp-1] != '\n' {
					ret = append(ret, src[p:p+pp+len(catalog.en)]...)
					p = p + pp + len(catalog.en)
					continue
				}
			}
		} else {
			// 一致後がpostでない場合はスキップ
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
			for range count {
				ret = append(ret, ' ')
			}
			ret = append(ret, hen...)
			if inCDATA(src[:p+pp]) {
				ret = append(ret, []byte("--><![CDATA[\n")...)
			} else {
				ret = append(ret, []byte("-->\n")...)
			}

			if catalog.ja != "" {
				for range count - 1 {
					ret = append(ret, ' ')
				}
				ret = append(ret, []byte(catalog.ja)...)
				ret = append(ret, src[p+pp+count+len(catalog.en):]...)
			} else {
				ret = append(ret, src[p+pp+count+len(catalog.en)+1:]...)
			}
			break
		}

		// Already in Japanese.
		ret = append(ret, src[p:p+pp+count+len(catalog.en)]...)
		p = p + pp + count + len(catalog.en)
	}
	return ret
}

func countLeadingSpaces(src []byte, pp int) int {
	count := 0
	for i := pp - 1; i >= 0; i-- {
		if src[i] == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

// 共通カタログを一つずつ置き換える
func matchCommon(src []byte, catalog Catalog) []byte {
	if !bytes.Contains(src, []byte(catalog.en)) {
		return src
	}
	// 正規表現にマッチした部分を置き換える
	src = catalog.enReg.ReplaceAllFunc(src, func(match []byte) []byte {
		space := ""
		if catalog.ja[0] == ' ' {
			space = leftPadSpace(string(match))
		}
		// 新しい日本語訳を追加
		en := REVHIGHHUN.ReplaceAll(match, []byte("&#45;-"))
		ret := "<!--\n" + string(en) + "-->\n" + space + catalog.ja + "\n"
		return []byte(ret)
	})
	return src
}

func leftPadSpace(str string) string {
	spaceCount := 0
	for i := 0; i < len(str); i++ {
		if str[i] == ' ' {
			spaceCount++
		} else {
			break
		}
	}
	space := ""
	if spaceCount > 1 {
		space = strings.Repeat(" ", spaceCount-1)
	}
	return space
}

func regCompile(catalogs Catalogs) Catalogs {
	for i := range catalogs {
		catalogs[i].enReg = regexp.MustCompile(`(?s)[^\n]*` + regexp.QuoteMeta(catalogs[i].en) + `\n`)
	}
	return catalogs
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
func matchAdditional(src []byte, catalog Catalog) []byte {
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
		if strings.Contains(catalog.ja, "split-") {
			break
		}
		// Already converted.
		p = p + i + j + len(catalog.pre) + 1
	}
	return -1
}

func (rep *Rep) unmatchedReplace(fileName string, src []byte) ([]byte, error) {
	ret := src
	// <para>ブロックの書き換え
	ret = rep.paraBlockReplace(ret)
	// 空カッコ《》の書き換え
	ret = BLANKBRACKET.ReplaceAllFunc(ret, rep.blankBracketReplace)
	return ret, nil
}

func (rep *Rep) paraBlockReplace(src []byte) []byte {
	blocks := splitBlock(src)
	for _, block := range blocks {
		blockSrc := string(block)
		ret := rep.blockReplace(blockSrc)
		if blockSrc == ret {
			continue
		}
		src = bytes.Replace(src, block, []byte(ret), 1)
	}
	return src
}

func (rep *Rep) blockReplace(src string) string {
	urlPost := ""
	cName := ""
	rSrc := src
	// <ulink url=\"&commit_baseurl 含まれていたらその前までを対象にする
	if idx := strings.Index(rSrc, "<ulink url=\"&commit_baseurl"); idx >= 0 {
		//urlPost = rSrc[idx:]
		rSrc = rSrc[:idx]
		// src内の最後の()を含む内容をcNameに入れる
		if submatches := regexp.MustCompile(`\([^)]*\)`).FindAllStringSubmatch(rSrc, -1); len(submatches) > 0 {
			cName = submatches[len(submatches)-1][0] // 最後のマッチを取得
		}
		// cName内の改行と連続スペースを一つのスペースに変換
		cName = strings.ReplaceAll(cName, "\n", " ")                   // 改行をスペースに変換
		cName = regexp.MustCompile(`\s+`).ReplaceAllString(cName, " ") // 連続スペースを一つのスペースに変換
		//log.Println("blockReplace:", cName, urlPost)
	}

	rSrc = strings.TrimLeft(rSrc, "\n")
	rSrc = strings.TrimRight(rSrc, "\n")
	srcBlock := strings.Split(rSrc, "\n")
	if len(srcBlock) < 3 {
		return src
	}
	b := 1
	for srcBlock[b] == "" {
		b += 1
	}
	pre := strings.Join(srcBlock[0:b], "\n")
	a := len(srcBlock) - 1
	for srcBlock[a-1] == "" {
		a -= 1
	}
	post := strings.Join(srcBlock[a:], "\n")
	body := strings.Join(srcBlock[b:a], "\n")
	enStr := stripNL(body)
	simJa, score := rep.findSimilar(enStr)

	if simJa == "no translation" {
		return src
	}

	// 類似文の日本語訳
	simJa = STRIPPMT.ReplaceAllString(simJa, "")
	simJa = STRIPM.ReplaceAllString(simJa, "")
	simJa = strings.TrimLeft(simJa, " ")
	simJa = strings.TrimRight(simJa, "\n")
	// 機械翻訳のためのマークを付ける
	mtStr := enStr
	if cName != "" {
		mtStr = mtStr[:len(mtStr)-len(cName)]
	}
	mtJa := rep.mtMark(mtStr, score)
	ret, err := replaceDst(score, simJa, mtJa)
	if err != nil {
		return src
	}

	org := REVHIGHHUN2.ReplaceAllString(body, "&#45;&#45;-")
	org = REVHIGHHUN.ReplaceAllString(org, "&#45;-")
	if cName != "" && !strings.HasSuffix(cName, "\n") {
		cName += "\n"
	}
	dst := fmt.Sprintf("%s\n<!--\n%s\n-->\n%s\n%s%s%s", pre, org, ret, cName, urlPost, post)
	return strings.Replace(src, rSrc, dst, 1)
}

func newTextra(apiConfig jpugDocConfig) (*textra.TexTra, error) {
	if apiConfig.APIKEY == "" || apiConfig.APISecret == "" {
		return nil, fmt.Errorf("textra: API KEY and API Secret are required")
	}
	config := textra.Config{}
	config.ClientID = apiConfig.APIKEY
	config.ClientSecret = apiConfig.APISecret
	config.Name = apiConfig.APIName
	return textra.New(config)
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

// srcを新しいcatalogの日本語訳に置き換えて更新する
func (rep Rep) updateFromCatalog(fileName string, src []byte) ([]byte, error) {
	srcDiff, err := getDiff(rep.vTag, fileName)
	if err != nil {
		return nil, err
	}
	org := Extraction(srcDiff)
	for _, o := range org {
		src = updateReplaceCatalog(src, rep.catalogs, o, rep.wip)
		src = updateReplaceCatalog(src, rep.common, o, rep.wip)
	}
	return src, nil
}

// enが一致してjaが違う場合は更新する
func updateReplaceCatalog(src []byte, catalogs []Catalog, o Catalog, wip bool) []byte {
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
	if ja == "" {
		return src
	}
	if len(o.ja) < 10 {
		return src
	}
	if !wip && STRIPM.Match([]byte(ja)) {
		return src
	}
	if strings.Contains(ja, "<title>") {
		return src
	}
	return bytes.ReplaceAll(src, []byte(o.ja), []byte(ja))
}

// <!-- -->《》に実際のマッチ度、機械翻訳を入れて置き換え
func (rep *Rep) blankBracketReplace(src []byte) []byte {
	matches := COMMENTSTART.FindAllIndex(src, -1)
	if len(matches) == 0 {
		return src
	}
	lastMatch := matches[len(matches)-1]
	subMatch := BLANKBRACKET2.FindSubmatch(src[lastMatch[1]:])
	if subMatch == nil {
		return src
	}
	enStr := stripNL(string(subMatch[1]))
	simJa, score := rep.findSimilar(enStr)
	if simJa == "no translation" {
		return src
	}
	simJa = STRIPPMT.ReplaceAllString(simJa, "")
	simJa = STRIPM.ReplaceAllString(simJa, "")
	simJa = strings.TrimLeft(simJa, " ")
	simJa = strings.TrimRight(simJa, "\n")
	// 機械翻訳のためのマークを付ける
	mtJa := rep.mtMark(enStr, score)
	ret, err := replaceDst(score, simJa, mtJa)
	if err != nil {
		log.Println(err.Error())
		return src
	}
	para := fmt.Sprintf("$1%s\n", ret)
	return BLANKBRACKET.ReplaceAll(src, []byte(para))
}

// 機械翻訳のためのマークを付ける
func (rep *Rep) mtMark(enStr string, score float64) string {
	if score < float64(rep.mt) {
		return MTTransStart + enStr + MTTransEnd
	}
	return ""
}

// 類似文、機械翻訳（マーク）へ置き換え
func (rep *Rep) simMtReplace(src []byte, pre string, org string, enStr string, post string) ([]byte, error) {
	simJa, score := rep.findSimilar(enStr)
	if simJa == "no translation" {
		return nil, nil
	}
	// <returnvalue>が含まれていたらスキップ
	if strings.Contains(enStr, "<returnvalue>") {
		return nil, nil
	}

	if Verbose {
		if simJa != "" {
			fmt.Printf("Similar...[%f][%.30s]\n", score, enStr)
		}
	}
	// 機械翻訳のためのマークを付ける
	mtJa := rep.mtMark(enStr, score)

	ej, err := replaceDst(score, simJa, mtJa)
	if err != nil {
		log.Println(err.Error())
		return nil, nil
	}
	para := fmt.Sprintf("$1%s<!--\n%s\n-->\n%s%s$3", pre, org, ej, post)
	return []byte(para), nil
}

func replaceDst(score float64, simJa string, mtJa string) (string, error) {
	switch {
	case simJa != "" && mtJa != "":
		return fmt.Sprintf("《マッチ度[%f]》%s\n《機械翻訳》%s", score, simJa, mtJa), nil
	case simJa != "":
		return fmt.Sprintf("《マッチ度[%f]》%s", score, simJa), nil
	case mtJa != "":
		return fmt.Sprintf("《機械翻訳》%s", mtJa), nil
	default:
		return "", fmt.Errorf("no match")
	}
}

func (rep *Rep) findSimilar(enStr string) (string, float64) {
	simJa, maxScore := findSimilar(rep.catalogs, enStr)
	if rep.similar == 0 || maxScore < float64(rep.similar) {
		return "", 0
	}
	// 機械翻訳の類似文を除外
	if strings.Contains(simJa, "《機械翻訳》") {
		return "", 0
	}
	// すでに類似文マークがある場合はマークを外す
	if strings.Contains(simJa, "《マッチ度") {
		simJa = strings.Split(simJa, "》")[1]
	}
	simJa = strings.TrimRight(simJa, "\n")

	return simJa, maxScore
}

// 類似文を探す
func findSimilar(catalogs []Catalog, enStr string) (string, float64) {
	var maxScore float64
	var simJa string
	// 最も類似度の高い日本語を探す
	for _, catalog := range catalogs {
		distance := levenshtein.ComputeDistance(enStr, catalog.en)
		score := (1 - (float64(distance) / float64(len(enStr)))) * 100
		if score > maxScore {
			simJa = catalog.ja
			maxScore = score
		}
	}
	return simJa, maxScore
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

/*
func (rep *Rep) replaceTitle(src []byte) []byte {
	idxes := TITLE2.FindAllSubmatchIndex(src, -1)
	ret := src
	for _, idx := range idxes {
		en := src[idx[2]:idx[3]]
		ja := src[idx[4]:idx[5]]
		if nja, ok := rep.titles[string(en)]; ok {
			if nja != string(ja) {
				ret = bytes.Replace(ret, ja, []byte(nja), -1)
			}
		}
	}

	idexes := TITLE.FindAllSubmatchIndex(src, -1)
	for _, idx := range idexes {
		en := src[idx[2]:idx[3]]
		ent := strings.TrimLeft(string(en), " ")
		spc := len(en) - len(ent)
		nja := ""
		if strings.Contains(ent, "Release ") {
			v := RELEASENUM.Find(en)
			if v != nil {
				nja = "<title>リリース" + string(v) + "</title>"
			}
		}

		if strings.Contains(ent, "Migration to Version") {
			v := RELEASENUM.Find(en)
			if v != nil {
				nja = "<title>バージョン" + string(v) + "への移行</title>"
			}
		}

		if j, ok := rep.titles[ent]; ok {
			nja = j
		}
		if len(nja) > 0 {
			title := fmt.Sprintf("\n<!--\n%s\n-->\n%s%s\n%s", en, strings.Repeat(" ", spc), nja, src[idx[4]:idx[5]])
			ret = bytes.Replace(ret, src[idx[0]:idx[1]], []byte(title), -1)
		}
	}
	return ret
}
*/
