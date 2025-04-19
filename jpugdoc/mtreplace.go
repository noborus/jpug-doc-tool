package jpugdoc

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/english"

	"github.com/noborus/go-textra"
)

type MTType struct {
	APIAutoTranslateType string
	cli                  *textra.TexTra
	tokenizer            *sentences.DefaultSentenceTokenizer
	maxTranslate         int
	count                int
}

var MTMARKREG = regexp.MustCompile(`«(.*?)»`)

func MTReplace(fileNames []string, limit int, prompt bool) error {
	cli, err := newTextra(Config)
	if err != nil {
		return fmt.Errorf("textra: %s", err)
	}
	tokenizer, err := english.NewSentenceTokenizer(nil)
	if err != nil {
		return err
	}
	mt := MTType{
		APIAutoTranslateType: Config.APIAutoTranslateType, // PostgreSQLマニュアル翻訳
		cli:                  cli,
		maxTranslate:         limit,
		tokenizer:            tokenizer,
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
		if mt.count > mt.maxTranslate {
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
		if mt.count > mt.maxTranslate {
			log.Println("Limit the number of requests")
			return m
		}
		// Remove the « and »
		englishText := m[MTL : len(m)-MTL]
		// Translate the text
		japaneseText, err := mt.machineTranslate(englishText)
		if err != nil {
			globalErr = err
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
	if mt.count > mt.maxTranslate {
		return MTTransStart + origin + MTTransEnd, nil
	}

	var origins []string
	if len(origin) > 4000 {
		origins = mt.splitSentences(origin)
	} else {
		origins = []string{origin}
	}

	var ja strings.Builder
	for _, origin := range origins {
		origin = strings.TrimRight(origin, ".") + "."
		translated, err := mt.cli.Translate(mt.APIAutoTranslateType, origin)
		if err != nil {
			log.Println("machineTranslate:", err)
			return "", fmt.Errorf("textra err: [%s] %w", origin, err)
		}
		ja.WriteString(translated)
	}
	return postProcess(ja.String()), nil
}

func (mt *MTType) splitSentences(src string) []string {
	sentences := mt.tokenizer.Tokenize(src)
	ret := make([]string, 0, len(sentences))
	for _, s := range sentences {
		ret = append(ret, s.Text)
	}
	return ret
}

func postProcess(str string) string {
	re := regexp.MustCompile(`\((.*?)\)`)
	str = re.ReplaceAllStringFunc(str, func(m string) string {
		// Remove the ( and )
		inner := m[1 : len(m)-1]
		// Check if the inner string contains any full-width characters
		for _, r := range inner {
			if r > '\u007F' {
				// If it does, replace the brackets with full-width brackets
				return "（" + inner + "）"
			}
		}
		// If it doesn't, return the match as is
		return m
	})
	str = KUTEN2.ReplaceAllStringFunc(str, func(m string) string {
		if m == "。）" {
			return "。）\n"
		}
		return "。\n"
	})
	return strings.TrimRight(str, "\n")
}
