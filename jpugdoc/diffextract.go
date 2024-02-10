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
	for i := 0; i < 3; i++ {
		if !scanner.Scan() {
			return
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
	var comment, jadd, extadd, indexF bool
	var addPre string

	reader := bytes.NewReader(diffSrc)
	scanner := bufio.NewScanner(reader)
	skipHeader(scanner)

	for scanner.Scan() {
		diffLine := scanner.Text()
		line := diffLine[1:]
		text := strings.TrimSpace(diffLine)
		extadd = false
		if !strings.HasPrefix(diffLine, "+") {
			prefixes = append(prefixes[1:], line)
		}
		// CDATA
		if m := STARTADDCOMMENTWITHC.FindAllStringSubmatch(text, 1); len(m) > 0 {
			catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA)
			en.Reset()
			ja.Reset()
			if len(m[0]) == 1 {
				en.WriteString("\n")
				comment = true
				continue
			}
			preCDATA = strings.Join(m[0][1:], "")
			comment = true
			continue
		} else if STARTADDCOMMENT.MatchString(text) {
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
			comment = true
			continue
		} else if ENDADDCOMMENT.MatchString(text) || ENDADDCOMMENTWITHC.MatchString(text) {
			comment = false
			jadd = true
			continue
		}

		if comment {
			if diffLine[0] == '-' {
				continue
			}
			line = REPHIGHHUN.ReplaceAllString(line, "-")
			en.WriteString(line)
			en.WriteString("\n")
		} else {
			if jadd && strings.HasPrefix(diffLine, "+") {
				ja.WriteString(line)
				ja.WriteString("\n")
			} else {
				jadd = false
			}
		}

		if comment || jadd {
			continue
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
					catalog := Catalog{
						pre: strings.Join(prefixes[:len(prefixes)-1], "\n") + "\n",
						ja:  line,
					}
					catalogs = append(catalogs, catalog)
					continue
				}
			}
			if indexF {
				indexj.WriteString(line)
				indexj.WriteString("\n")
			}
			if !indexF && !jadd {
				if !strings.Contains(diffLine, "</indexterm>") {
					if strings.Join(prefixes, "") != "" {
						if addja.Len() == 0 {
							addPre = strings.Join(prefixes, "\n") + "\n"
						}
						extadd = true
						addja.WriteString(line)
						addja.WriteString("\n")
					}
				}
			}
		}
		if !extadd && addja.Len() != 0 {
			catalogs = addJaCatalogs(catalogs, addPre, addja)
			addja.Reset()
		}
	}
	// last
	catalogs = addCatalogs(catalogs, prefix, en, ja, preCDATA)

	return catalogs
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
		ja:  strings.Trim(ja.String(), "\n"),
	}
	if ja.Len() != 0 {
		catalogs = append(catalogs, catalog)
	}
	return catalogs
}
