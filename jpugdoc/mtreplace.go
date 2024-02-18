package jpugdoc

import (
	"fmt"
	"log"
	"regexp"

	"github.com/noborus/go-textra"
)

type MTType struct {
	APIAutoTranslateType string
	cli                  *textra.TexTra
	count                int
}

var MTMARKREG = regexp.MustCompile(`«(.*?)»`)

var MaxTranslate = 100

func MTReplace(fileNames []string, prompt bool) error {
	cli, err := newTextra(Config)
	if err != nil {
		return fmt.Errorf("textra: %s", err)
	}
	var mt = MTType{
		APIAutoTranslateType: Config.APIAutoTranslateType, // PostgreSQLマニュアル翻訳
		cli:                  cli,
	}

	for _, fileName := range fileNames {
		if Verbose {
			log.Println("mt:", fileName)
		}
		s, err := ReadAllFile(fileName)
		if err != nil {
			log.Printf("skip:%s:%s\n", fileName, err)
			return nil
		}
		src := string(s)
		ret, err := mt.Replace(src, prompt)
		if err != nil {
			return fmt.Errorf("replace: %s: %w", fileName, err)
		}
		if len(ret) == 0 || src == ret {
			continue
		}
		if err := rewriteFile(fileName, []byte(ret)); err != nil {
			return fmt.Errorf("rewrite: %s: %w", fileName, err)
		}
		fmt.Printf("MTreplace: %s\n", fileName)
		if mt.count > MaxTranslate {
			break
		}
	}
	return nil
}

func (mt *MTType) Replace(src string, prompt bool) (string, error) {
	return mt.ReplaceText(src)
}

func (mt *MTType) ReplaceText(src string) (string, error) {
	var globalErr error
	ret := MTMARKREG.ReplaceAllStringFunc(src, func(m string) string {
		// Remove the « and »
		englishText := m[MTL : len(m)-MTL]
		// Translate the text
		japaneseText, err := mt.machineTranslate(englishText)
		if err != nil {
			globalErr = err
			return m
		}
		if mt.count > MaxTranslate {
			log.Println("Limit the number of requests")
			return m
		}
		// Return the translated text
		return japaneseText
	})
	return ret, globalErr
}

func (mt *MTType) machineTranslate(origin string) (string, error) {
	if Verbose {
		log.Println("machineTranslate:", origin)
	}
	// Limit the number of requests
	mt.count++
	if mt.count > MaxTranslate {
		return MTTransStart + origin + MTTransEnd, nil
	}
	ja, err := mt.cli.Translate(mt.APIAutoTranslateType, string(origin))
	if err != nil {
		log.Println("machineTranslate:", err)
		return "", err
	}
	return ja, nil
}
