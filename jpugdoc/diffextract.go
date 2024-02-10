package jpugdoc

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os/exec"
	"strings"
)

// ファイル名の配列を受け取り、それぞれのファイル名のdiffから原文と日本語訳の対の配列を抽出し、
// それぞれのファイル名に対応するカタログファイル(filename.sgml.t)を作成する
func Extract(fileNames []string) {
	vTag, err := versionTag()
	if err != nil {
		log.Fatal(err)
	}

	for _, fileName := range fileNames {
		diffSrc := getDiff(vTag, fileName)
		catalogs := Extraction(diffSrc)
		saveCatalog(fileName, catalogs)
	}
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
func getDiff(vTag string, fileName string) []byte {
	// git diff --histogram -U10000 REL_16_0 doc/src/sgml/ref/backup.sgml
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
		log.Fatal("getDiff", err)
	}
	cmd.Wait()
	return src
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

	prefixes := make([]string, 10)
	prefix := ""
	preCDATA := ""
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
			catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA)
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
				if !strings.HasSuffix(ja.String(), ");\n") {
					ja.WriteString(");\n")
				}
			}
			catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA)
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
				continue
			}
			// 日本語の追加が終了
			japaneseF = false
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
						catalogs = addJaCatalogs(catalogs, index.String(), indexj)
					}
					index.Reset()
					indexj.Reset()
				}
			} else if ENDINDEXTERM.MatchString(text) {
				indexj.WriteString(line)
				indexj.WriteString("\n")
				indexF = false
				if index.Len() > 0 {
					catalogs = addJaCatalogs(catalogs, index.String(), indexj)
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
			catalogs = addJaCatalogs(catalogs, addPre, addja)
			addja.Reset()
		}
	}
	// last
	catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA)
	catalogs = addJaCatalogs(catalogs, addPre, addja)
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

func addCatalogs(catalogs []Catalog, pre string, en strings.Builder, ja strings.Builder, preCDATA string) []Catalog {
	catalog := Catalog{
		pre:      pre,
		en:       strings.Trim(en.String(), "\n"),
		ja:       strings.Trim(ja.String(), "\n"),
		preCDATA: preCDATA,
	}
	if en.Len() != 0 {
		catalogs = append(catalogs, catalog)
	}
	return catalogs
}

func addJaCatalogs(catalogs []Catalog, pre string, ja strings.Builder) []Catalog {
	catalog := Catalog{
		pre: pre,
		ja:  strings.TrimSuffix(ja.String(), "\n"),
	}
	if ja.Len() != 0 {
		catalogs = append(catalogs, catalog)
	}
	return catalogs
}
