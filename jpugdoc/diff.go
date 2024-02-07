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

	pre := make([]string, 10)
	prefix := ""
	cdatapre := ""
	var catalogs []Catalog
	var comment, jadd, extadd, indexF bool
	var addPre string

	reader := bytes.NewReader(diffSrc)
	scanner := bufio.NewScanner(reader)
	skipHeader(scanner)

	for scanner.Scan() {
		l := scanner.Text()
		line := strings.TrimSpace(l)
		extadd = false
		pre = append(pre[1:], l[1:])
		// CDATA
		if m := STARTADDCOMMENTWITHC.FindAllStringSubmatch(line, 1); len(m) > 0 {
			catalog := Catalog{
				pre:      prefix,
				en:       strings.Trim(en.String(), "\n"),
				ja:       strings.Trim(ja.String(), "\n"),
				cdatapre: cdatapre,
			}
			if en.Len() != 0 {
				catalogs = append(catalogs, catalog)
			}
			en.Reset()
			ja.Reset()
			prefix = pre[0]
			if len(m[0]) == 1 {
				en.WriteString("\n")
				comment = true
				continue
			}
			cdatapre = strings.Join(m[0][1:], "")
			//en.WriteString(cdatapre)
			//en.WriteString("\n")
			comment = true
			continue
		} else if STARTADDCOMMENT.MatchString(line) {
			if strings.HasSuffix(en.String(), "\n);\n") {
				if !strings.HasSuffix(ja.String(), ");\n") {
					ja.WriteString(");\n")
				}
			}
			catalog := Catalog{
				pre: prefix,
				en:  strings.Trim(en.String(), "\n"),
				ja:  strings.Trim(ja.String(), "\n"),
			}
			if en.Len() != 0 {
				catalogs = append(catalogs, catalog)
			}
			en.Reset()
			ja.Reset()
			prefix = pre[0]
			en.WriteString("\n")
			comment = true
			continue
		} else if ENDADDCOMMENT.MatchString(line) || ENDADDCOMMENTWITHC.MatchString(line) {
			comment = false
			jadd = true
			continue
		}

		if comment {
			if l[0] == '-' {
				continue
			}
			l = REPHIGHHUN.ReplaceAllString(l, "-")
			en.WriteString(l[1:])
			en.WriteString("\n")
		} else {
			if jadd && strings.HasPrefix(l, "+") {
				ja.WriteString(l[1:])
				ja.WriteString("\n")
			} else {
				jadd = false
			}
		}

		if comment {
			continue
		}

		// indexterm,etc.
		if !strings.HasPrefix(l, "+") {
			// original
			if STARTINDEXTERM.MatchString(line) {
				index.Reset()
				indexF = true
				if ENDINDEXTERM.MatchString(line) {
					index.WriteString(l[1:])
					index.WriteString("\n")
					indexF = false
				}
			} else if ENDINDEXTERM.MatchString(line) {
				index.WriteString(l[1:])
				index.WriteString("\n")
				indexF = false
			}
			if indexF {
				index.WriteString(l[1:])
				index.WriteString("\n")
			}
		} else {
			// Add ja indexterm
			if STARTINDEXTERM.MatchString(line) {
				indexF = true
				if ENDINDEXTERM.MatchString(line) {
					indexj.WriteString(l[1:])
					indexj.WriteString("\n")
					indexF = false
					if index.Len() > 0 {
						catalog := Catalog{
							pre: index.String(),
							ja:  strings.Trim(indexj.String(), "\n"),
						}
						catalogs = append(catalogs, catalog)
					}
					index.Reset()
					indexj.Reset()
				}
			} else if ENDINDEXTERM.MatchString(line) {
				indexj.WriteString(l[1:])
				indexj.WriteString("\n")
				indexF = false
				if index.Len() > 0 {
					catalog := Catalog{
						pre: index.String(),
						ja:  strings.Trim(indexj.String(), "\n"),
					}
					catalogs = append(catalogs, catalog)
				}
				index.Reset()
				indexj.Reset()
			} else {
				if SPLITCOMMENT.MatchString(line) {
					catalog := Catalog{
						pre: strings.Join(pre[:len(pre)-1], "\n") + "\n",
						ja:  l[1:],
					}
					catalogs = append(catalogs, catalog)
					continue
				}
			}
			if indexF {
				indexj.WriteString(l[1:])
				indexj.WriteString("\n")
			}
			if !indexF && !jadd {
				if !strings.Contains(l, "</indexterm>") {
					if strings.Join(pre[:len(pre)-1], "") != "" {
						if addja.Len() == 0 {
							addPre = strings.Join(pre[:len(pre)-1], "\n") + "\n"
						}
						extadd = true
						addja.WriteString(l[1:])
						addja.WriteString("\n")
					}
				}
			}
		}
		if !extadd && addja.Len() != 0 {
			catalog := Catalog{
				pre: addPre,
				ja:  strings.Trim(addja.String(), "\n"),
			}
			catalogs = append(catalogs, catalog)
			addja.Reset()
		}
	}
	// last
	if en.Len() != 0 {
		catalog := Catalog{
			pre: prefix,
			en:  strings.Trim(en.String(), "\n"),
			ja:  strings.Trim(ja.String(), "\n"),
		}
		catalogs = append(catalogs, catalog)
	}

	return catalogs
}
