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
	catalogs []Catalog
	titles   map[string]string
	vTag     string
	update   bool
	similar  int
	mt       int
	wip      bool
	prompt   bool
}

// Replace は指定されたファイル名のファイルを置き換える
func Replace(fileNames []string, vTag string, update bool, similar int, mt int, wip bool, prompt bool) error {
	rep, err := newOpts(vTag, update, similar, mt, prompt)
	if err != nil {
		return err
	}
	rep.titles = titleMap()
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

func newOpts(vTag string, update bool, similar int, mt int, prompt bool) (*Rep, error) {
	rep := &Rep{
		similar: similar,
		update:  update,
		mt:      mt,
	}
	if update && vTag == "" {
		v, err := versionTag()
		if err != nil {
			return rep, err
		}
		rep.vTag = v
	}
	rep.prompt = prompt

	return rep, nil
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
	src = rep.replaceTitle(src)
	// 追加形式の翻訳文を追加
	for _, catalog := range rep.catalogs {
		if catalog.en == "" {
			src = rep.matchAdditional(src, catalog)
		}
	}
	// コメント形式の翻訳文を追加
	for _, catalog := range rep.catalogs {
		if catalog.en != "" {
			src = rep.matchComment(src, catalog)
		}
	}
	return src
}

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

// カタログを一つづつ置き換える
func (rep Rep) matchComment(src []byte, catalog Catalog) []byte {
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
func (rep Rep) matchAdditional(src []byte, catalog Catalog) []byte {
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
	block := bytes.Buffer{}
	ret := bytes.Buffer{}
	para := bytes.NewBuffer(src)
	tag := "none"
	for {
		line, err := para.ReadString('\n')
		if err != nil {
			r := rep.blockReplace(tag, block.String())
			ret.WriteString(r)
			break
		}
		if sub := REPARABLOCK.FindAllStringSubmatch(line, 1); sub != nil {
			preTag := sub[0][1]
			if len(preTag) < 30 {
				r := rep.blockReplace(tag, block.String())
				ret.WriteString(r)
				block.Reset()
				tag = preTag
			}
		}
		block.WriteString(line)
	}
	return ret.Bytes()
}

func (rep *Rep) blockReplace(tag string, src string) string {
	str := REPARABLOCK.ReplaceAllString(src, "")
	if strings.TrimSpace(str) == "" {
		return src
	}
	if tag == "programlisting" || tag == "screen" || tag == "synopsis" || tag == "/varlistentry" {
		return src
	}
	if tag != "para" && !strings.HasPrefix(tag, "/programlisting") && !strings.HasPrefix(tag, "/screen") && !strings.HasPrefix(tag, "/synopsis") {
		return src
	}
	rSrc := removeEmptyLines(str)
	// 既に翻訳済みの場合はスキップ
	if strings.Contains(rSrc, "<!--") || strings.Contains(rSrc, "-->") {
		return src
	}
	// <para>が含まれている場合は改行されていないのでスキップ
	if strings.Contains(rSrc, "<para>") || strings.Contains(rSrc, "</para>") {
		return src
	}
	if strings.Contains(rSrc, "<title>") {
		return src
	}
	// <returnvalue>が含まれていたらスキップ
	if strings.Contains(rSrc, "<returnvalue>") {
		return src
	}

	if NIHONGO.MatchString(str) {
		return src
	}
	d := removeEmptyLines(str)
	if d == "" {
		return src
	}
	enStr := stripNL(rSrc)
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
	dst := fmt.Sprintf("<!--\n%s-->\n%s\n", rSrc, ret)
	return strings.Replace(src, rSrc, dst, 1)
}

func removeEmptyLines(s string) string {
	re := regexp.MustCompile(`(?m)^\n+`)
	return re.ReplaceAllString(s, "")
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

/*
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
*/

// srcを新しいcatalogの日本語訳に置き換えて更新する
func (rep Rep) updateFromCatalog(fileName string, src []byte) ([]byte, error) {
	srcDiff, err := getDiff(rep.vTag, fileName)
	if err != nil {
		return nil, err
	}
	org := Extraction(srcDiff)
	for _, o := range org {
		src = updateReplaceCatalog(src, rep.catalogs, o, rep.wip)
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
	return bytes.ReplaceAll(src, []byte(o.ja), []byte(ja))
}

// <para><screen>の置き換え
// ReplaceAllFuncで呼び出される
func (rep *Rep) paraScreenReplace(src []byte) []byte {
	subMatch := REPARASCREEN.FindSubmatch(src)
	// <para>\nで改行されていない場合はスキップ
	if !bytes.HasPrefix(subMatch[1][6:], []byte("\n")) {
		return src
	}
	// <screen>、<programlisting>、<synopsis>が含まれていない場合はスキップ
	if !containsAny(subMatch[3], [][]byte{[]byte("<screen>"), []byte("<programlisting>"), []byte("<synopsis>")}) {
		return src
	}

	para := subMatch[2]
	if len(bytes.TrimSpace(para)) == 0 {
		return src
	}
	// 既に翻訳済みの場合はスキップ
	if bytes.Contains(para, []byte("<!--")) {
		return src
	}

	org := strings.TrimRight(string(para), "\n")
	org = REVHIGHHUN2.ReplaceAllString(org, "&#45;&#45;-")
	org = REVHIGHHUN.ReplaceAllString(org, "&#45;-")
	stripOrg := stripNL(string(para))
	// 類似文、機械翻訳置き換え
	ret, err := rep.simMtReplace(src, "", org, stripOrg, "")
	if err != nil {
		log.Println(err.Error())
	}
	if ret == nil {
		return src
	}
	if rep.prompt {
		return promptReplace(src, ret)
	}
	return REPARASCREEN.ReplaceAll(src, ret)
}

func containsAny(b []byte, subs [][]byte) bool {
	for _, sub := range subs {
		if bytes.Contains(b, sub) {
			return true
		}
	}
	return false
}

// <para></para>の置き換え
// ReplaceAllFuncで呼び出される
func (rep *Rep) paraReplace(src []byte) []byte {
	subMatch := REPARA.FindSubmatch(src)
	tag := string(subMatch[1])
	para := subMatch[2]
	// </para>の前に改行がない場合はスキップ
	if !bytes.Contains(subMatch[3], []byte("\n")) {
		return src
	}
	org := strings.TrimRight(string(para), "\n")
	org = REVHIGHHUN2.ReplaceAllString(org, "&#45;&#45;-")
	org = REVHIGHHUN.ReplaceAllString(org, "&#45;-")
	stripOrg := stripNL(string(para))
	// <para>\nで改行されていない場合はスキップ
	if !strings.Contains(tag, "\n") {
		return src
	}
	// 既に翻訳済みの場合はスキップ
	if strings.HasPrefix(stripOrg, "<!--") {
		return src
	}
	// 既に翻訳済みの場合はスキップ
	if strings.Contains(stripOrg, "<para>") && strings.Contains(stripOrg, "<!--") {
		return src
	}
	// 空白のみの場合はスキップ
	if strings.Trim(stripOrg, " ") == "" {
		return src
	}
	// 既に日本語がある場合はスキップ
	if NIHONGO.MatchString(stripOrg) {
		return src
	}
	pre, org, stripOrg, post, err := lastBlock(para, org, stripOrg)
	if err != nil {
		return src
	}
	// <returnvalue>,<screen>が含まれている場合はスキップ
	if strings.Contains(org, "<returnvalue>") || strings.Contains(org, "<screen>") || strings.Contains(org, "<programlisting>") {
		return src
	}
	// 翻訳不要の場合はスキップ
	for _, catalog := range rep.catalogs {
		if catalog.ja == "no translation" {
			if catalog.en == stripOrg {
				if Verbose {
					log.Printf("skip: %s\n", stripOrg)
				}
				return src
			}
		}
	}
	// 類似文、機械翻訳置き換え
	ret, err := rep.simMtReplace(src, pre, org, stripOrg, post)
	if err != nil {
		log.Println(err.Error())
	}
	if ret == nil {
		return src
	}
	if rep.prompt {
		return promptReplace(src, ret)
	}

	return REPARA.ReplaceAll(src, ret)
}

// para内の最後のブロックを返す
func lastBlock(para []byte, org string, stripOrg string) (string, string, string, string, error) {
	pre := ""
	post := ""
	p := 0
	// <para>が含まれている場合は次の<para>を置き換え
	if strings.Contains(stripOrg, "<para>") {
		l := bytes.Index(para, []byte("<para>"))
		// <para>の後が改行で終わっていない場合はスキップ
		if !bytes.HasPrefix(para[l+6:], []byte("\n")) {
			return "", "", "", "", fmt.Errorf("no para")
		}
		p = l + len("<para>\n")
	} else if strings.Contains(stripOrg, "</screen>") {
		// </screen>が含まれている場合は</screen>の後を置き換え
		p = bytes.LastIndex(para, []byte("</screen>")) + len("</screen>\n")
	} else if strings.Contains(stripOrg, "</programlisting>") {
		// </programlisting>が含まれている場合は</programlisting>の後を置き換え
		p = bytes.LastIndex(para, []byte("</programlisting>")) + len("</programlisting>\n")
	} else {
		return pre, org, stripOrg, post, nil
	}

	if p > len(para) {
		return "", "", "", "", fmt.Errorf("no para")
	}

	pre = string(para[:p])
	para = para[p:]
	org = string(para)
	stripOrg = stripNL(string(para))
	return pre, org, stripOrg, post, nil
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
