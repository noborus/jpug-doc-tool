package jpugdoc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ファイル名の配列を受け取り、それぞれのファイル名のdiffから原文と日本語訳の対の配列を抽出し、
// それぞれのファイル名に対応するカタログファイル(filename.sgml.t)を作成する
func Extract(vTag string, fileNames []string) error {
	if vTag == "" {
		v, err := versionTag()
		if err != nil {
			return err
		}
		vTag = v
	}

	for _, fileName := range fileNames {
		if Verbose {
			log.Printf("Extract: %s\n", fileName)
		}
		diffSrc, err := getDiff(vTag, fileName)
		if err != nil {
			return fmt.Errorf("getDiff: %w", err)
		}
		catalogs := Extraction(diffSrc)
		catalogs, err = noTransPara(catalogs, fileName)
		if err != nil {
			log.Println(err)
		}
		saveCatalog(fileName, catalogs)
	}
	return nil
}

// skip diff header
func skipHeader(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "@@") {
			break
		}
	}
}

// git diffの結果を省略せずに取得する
func getDiff(vTag string, fileName string) ([]byte, error) {
	// git diff --histogram -U10000 REL_16_0 doc/src/sgml/ref/backup.sgml
	args := []string{"diff", "--histogram", "-U10000", vTag, fileName}
	cmd := exec.Command("git", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var src []byte
	cmd.Start()
	src, err = io.ReadAll(stdout)
	if err != nil {
		log.Fatal("getDiff", err)
	}
	cmd.Wait()
	return src, nil
}

// diff を原文と日本語訳の対(Catalog)の配列に変換する
// <para>
// +<!--
// english
// +-->
// +japanese
// </para>
func Extraction(diffSrc []byte) []Catalog {
	var en, ja, addja, index, indexj strings.Builder

	prefixes := make([]string, 5)
	prefix := ""
	preCDATA := ""
	postfix := ""
	var catalogs []Catalog
	var englishF, japaneseF, addExtraF, indexF bool
	var addPre string

	reader := bytes.NewReader(diffSrc)
	scanner := bufio.NewScanner(reader)
	skipHeader(scanner)

	for scanner.Scan() {
		diffLine := scanner.Text()
		line := diffLine[1:]
		text := strings.TrimSpace(diffLine)
		addExtraF = false
		if !strings.HasPrefix(diffLine, "+") {
			prefixes = append(prefixes[1:], line)
		}

		if m := STARTADDCOMMENTWITHC.FindAllStringSubmatch(text, 1); len(m) > 0 { // CDATA
			catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA, postfix)
			en.Reset()
			ja.Reset()
			if len(m[0]) == 1 {
				en.WriteString("\n")
				englishF = true
				continue
			}
			preCDATA = strings.Join(m[0][1:], "")
			englishF = true
			continue
		} else if STARTADDCOMMENT.MatchString(text) { // <!--コメント始まり
			if strings.HasSuffix(en.String(), "\n);\n") {
				if !strings.HasSuffix(ja.String(), ");\n") { // ");"だけの行はdiffで英語、日本語、");"の順になってしまうので、補正する
					ja.WriteString(");\n")
				}
			}
			catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA, postfix)
			en.Reset()
			ja.Reset()
			prefix = prefixes[len(prefixes)-1]
			en.WriteString("\n")
			englishF = true
			continue
		} else if ENDADDCOMMENT.MatchString(text) || ENDADDCOMMENTWITHC.MatchString(text) { // コメント終わり
			englishF = false
			japaneseF = true
			continue
		}

		if englishF {
			/* 原文の`--`の置き換えによる ^- ^+ の差分は ^-を無視して^+を`--`に置き換えて追加する */
			if diffLine[0] == '-' {
				continue
			}
			en.WriteString(REPHIGHHUN.ReplaceAllString(line, "-"))
			en.WriteString("\n")
			continue
		}

		if japaneseF {
			if strings.HasPrefix(diffLine, "+") {
				ja.WriteString(line)
				ja.WriteString("\n")
				postfix = ""
				continue
			}
			// 日本語の追加が終了
			japaneseF = false
			postfix = line
		}

		// indexterm,etc.
		if !strings.HasPrefix(diffLine, "+") {
			// original
			if STARTINDEXTERM.MatchString(text) {
				index.Reset()
				indexF = true
				if ENDINDEXTERM.MatchString(text) {
					index.WriteString(line)
					index.WriteString("\n")
					indexF = false
				}
			} else if ENDINDEXTERM.MatchString(text) {
				index.WriteString(line)
				index.WriteString("\n")
				indexF = false
			}
			if indexF {
				index.WriteString(line)
				index.WriteString("\n")
			}
		} else {
			// Add ja indexterm
			if STARTINDEXTERM.MatchString(text) {
				indexF = true
				if ENDINDEXTERM.MatchString(text) {
					indexj.WriteString(line)
					indexj.WriteString("\n")
					indexF = false
					if index.Len() > 0 {
						catalogs = addJaCatalogs(catalogs, index.String(), indexj, postfix)
					}
					index.Reset()
					indexj.Reset()
				}
			} else if ENDINDEXTERM.MatchString(text) {
				indexj.WriteString(line)
				indexj.WriteString("\n")
				indexF = false
				if index.Len() > 0 {
					catalogs = addJaCatalogs(catalogs, index.String(), indexj, postfix)
				}
				index.Reset()
				indexj.Reset()
			} else {
				if SPLITCOMMENT.MatchString(text) {
					if strings.HasPrefix(diffLine, "+") {
						addPre = strings.Join(prefixBlock(prefixes), "\n") + "\n"
						addja.WriteString(line)
						addja.WriteString("\n")
					}
					continue
				}
			}
			if indexF {
				indexj.WriteString(line)
				indexj.WriteString("\n")
			}
			if !indexF && !japaneseF {
				if !strings.Contains(diffLine, "</indexterm>") {
					if strings.Join(prefixes, "") != "" {
						if addja.Len() == 0 {
							addPre = strings.Join(prefixBlock(prefixes), "\n") + "\n"
						}
						addExtraF = true
						addja.WriteString(line)
						addja.WriteString("\n")
					}
				}
			}
		}
		if !addExtraF && addja.Len() != 0 {
			catalogs = addJaCatalogs(catalogs, addPre, addja, line)
			addja.Reset()
		}
	}
	// last
	catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA, postfix)
	catalogs = addJaCatalogs(catalogs, addPre, addja, "")
	return catalogs
}

// 逆から最初のブロックを残す
func prefixBlock(s []string) []string {
	blockF := false
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == "" && i < len(s)-3 { // 最低3行は残す
			if blockF {
				return s[i+1:]
			}
		} else {
			blockF = true
		}
	}
	return s
}

func splitBlock(src []byte) [][]byte {
	block := bytes.Buffer{}
	ret := [][]byte{}
	srcBuff := bytes.NewBuffer(src)
	tag := "none"
	for {
		line, err := srcBuff.ReadString('\n')
		if err != nil {
			block.WriteString(line)
			r := block.String()
			if len(r) > 0 {
				ret = appendBlock(ret, tag, []byte(r))
			}
			break
		}
		if sub := RETAGBLOCK.FindAllStringSubmatch(line, 1); sub != nil {
			preTag := sub[0][1]
			if len(preTag) < 50 {
				block.WriteString(line)
				r := block.String()
				str := RETAGBLOCK.ReplaceAllString(r, "")
				if len(str) > 0 {
					// <para が含まれていたら<para〜以降を切り出す
					if idx := strings.Index(r, "<para "); idx > 0 {
						tag = "para"
						r = r[idx:]
					}
					ret = appendBlock(ret, tag, []byte(r))
				}
				block.Reset()
				tag = preTag
				block.WriteString(line)
				continue
			}
		}
		if CLOSEPARA.MatchString(line) {
			block.WriteString(line)
			r := block.String()
			para := REPARA2.FindAllStringSubmatch(r, -1)
			if para != nil {
				p := para[len(para)-1][1]
				p = strings.TrimRight(p, " ")
				if len(p) > 0 {
					ret = appendBlock(ret, "para", []byte(r))
				}
				block.Reset()
			}
			tag = "none"
			block.WriteString(line)
			continue
		}
		block.WriteString(line)
	}
	return ret
}

// 翻訳が必要な場合はtrueを返す
func isTranslate(tag string, src string) bool {
	if tag == "programlisting" || tag == "screen" || tag == "synopsis" || tag == "/varlistentry" {
		return false
	}
	if tag != "none" && tag != "para" && tag != "/indexterm" && !strings.HasPrefix(tag, "/programlisting") && !strings.HasPrefix(tag, "/screen") && !strings.HasPrefix(tag, "/synopsis") {
		return false
	}
	// 既に翻訳済みの場合はスキップ
	if strings.Contains(src, "<!--") || strings.Contains(src, "-->") {
		return false
	}

	if strings.Contains(src, "<title>") {
		return false
	}
	// <returnvalue>が含まれていたらスキップ
	if strings.Contains(src, "<returnvalue>") {
		return false
	}
	// <footnote>が含まれていたらスキップ
	if strings.Contains(src, "<footnote>") {
		return false
	}

	// str := RETAGBLOCK.ReplaceAllString(src, "")
	lines := strings.Split(src, "\n")
	if len(lines) < 3 {
		return false
	}
	body := strings.Join(lines[1:len(lines)-2], "\n")
	if !containsLetter(body) {
		return false
	}
	if NIHONGO.MatchString(body) {
		return false
	}
	return true
}

// containsLetter checks if the string contains any letter characters (a-z, A-Z).
func containsLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func appendBlock(blocks [][]byte, tag string, block []byte) [][]byte {
	if n := bytes.Trim(block, "\n"); len(n) == 0 {
		return blocks
	}
	blockSrc := string(block)
	if !isTranslate(tag, blockSrc) {
		return blocks
	}
	blocks = append(blocks, block)
	return blocks
}

func noTransPara(catalogs []Catalog, fileName string) ([]Catalog, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return catalogs, fmt.Errorf("noTransPara: %w", err)
	}
	defer f.Close()
	src, err := io.ReadAll(f)
	if err != nil {
		return catalogs, fmt.Errorf("noTransPara: %w", err)
	}

	paras := splitBlock(src)
	for _, para := range paras {
		lines := strings.Split(string(para), "\n")
		body := string(para)
		if len(lines) > 3 {
			body = strings.Join(lines[1:len(lines)-2], "\n")
		}
		paraStr := stripNL(body)
		// 既に翻訳済みの場合はスキップ
		if strings.HasPrefix(paraStr, "<!--") {
			continue
		}
		if paraStr == "" {
			continue
		}
		var en, ja strings.Builder
		en.WriteString(paraStr)
		ja.WriteString("no translation")
		catalogs = addCatalogs(catalogs, "", en, ja, "", "")
	}
	return catalogs, nil
}

func addCatalogs(catalogs []Catalog, pre string, en strings.Builder, ja strings.Builder, preCDATA string, post string) []Catalog {
	enStr := strings.Trim(en.String(), "\n")
	jaStr := strings.Trim(ja.String(), "\n")
	if enStr == "" && jaStr == "" {
		return catalogs
	}

	if post == enStr {
		post = ""
	}
	if post == ");" {
		post = ""
	}
	catalog := Catalog{
		pre:      pre,
		en:       enStr,
		ja:       jaStr,
		preCDATA: preCDATA,
		post:     post,
	}
	catalogs = append(catalogs, catalog)
	return catalogs
}

func addJaCatalogs(catalogs []Catalog, pre string, ja strings.Builder, post string) []Catalog {
	if ja.Len() == 0 {
		return catalogs
	}
	catalog := Catalog{
		pre:  pre,
		ja:   strings.TrimSuffix(ja.String(), "\n"),
		post: post,
	}
	catalogs = append(catalogs, catalog)
	return catalogs
}
