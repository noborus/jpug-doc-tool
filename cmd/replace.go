package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/clientcredentials"
)

type Catalog map[string]string
type Rep struct {
	catalog Catalog
	mt      bool
	similar int
}

const APIURL = "https://mt-auto-minhon-mlt.ucri.jgn-x.jp/"

func loadCatalog(fileName string) Catalog {
	src, err := ReadFile(fileName)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return nil
	}
	catalog := make(map[string]string)

	catas := SPLITCATALOG.FindAll(src, -1)
	for _, cata := range catas {
		re := SPLITCATALOG.FindSubmatch(cata)
		en := string(re[1])
		ja := string(re[2])
		catalog[en] = ja
	}
	return catalog
}

func (rep Rep) Replace(src []byte) []byte {
	ret := rep.paraReplace(src)
	return ret
}

type TextraResult struct {
	Resultset struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Request struct {
			URL   string `json:"url"`
			Text  string `json:"text"`
			Split int    `json:"split"`
			Data  string `json:"data"`
		} `json:"request"`
		Result struct {
			Text        string `json:"text"`
			Information struct {
				TextS    string `json:"text-s"`
				TextT    string `json:"text-t"`
				Sentence []struct {
					TextS string `json:"text-s"`
					TextT string `json:"text-t"`
					Split []struct {
						TextS   string `json:"text-s"`
						TextT   string `json:"text-t"`
						Process struct {
							Regex         []interface{} `json:"regex"`
							ReplaceBefore []interface{} `json:"replace-before"`
							Preprocess    []interface{} `json:"preprocess"`
							Translate     struct {
								Reverse       []interface{}     `json:"reverse"`
								Specification []interface{}     `json:"specification"`
								TextS         string            `json:"text-s"`
								TextT         string            `json:"text-t"`
								Associate     [][]interface{}   `json:"associate"`
								Oov           interface{}       `json:"oov"`
								Exception     string            `json:"exception"`
								Associates    [][][]interface{} `json:"associates"`
							} `json:"translate"`
							ReplaceAfter []interface{} `json:"replace-after"`
						} `json:"process"`
					} `json:"split"`
				} `json:"sentence"`
			} `json:"information"`
		} `json:"result"`
	} `json:"resultset"`
}

func textraTranslate(enstr string) string {
	ctx := context.Background()
	conf := &clientcredentials.Config{
		ClientID:     Conf.ClientID,
		ClientSecret: Conf.ClientSecret,
		TokenURL:     APIURL + "oauth2/token.php",
	}

	client := conf.Client(ctx)
	token, err := conf.Token(ctx)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%#v\n", token)
	values := url.Values{
		"access_token": []string{token.AccessToken},
		"key":          []string{Conf.ClientID},
		"api_name":     []string{Conf.APIName},
		"api_param":    []string{Conf.APIParam},
		"name":         []string{Conf.Name},
		"type":         []string{"json"},
		"text":         []string{enstr},
	}
	//fmt.Println(values.Encode())

	resp, err := client.PostForm(APIURL+"api/", values)
	if err != nil {
		log.Fatal(err)
	}
	//dumpResp, _ := httputil.DumpResponse(resp, true)
	//fmt.Printf("%s", dumpResp)
	s, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	data := new(TextraResult)
	if err := json.Unmarshal(s, data); err != nil {
		log.Fatal(err)
	}
	return data.Resultset.Result.Text
}

func (rep Rep) paraReplace(src []byte) []byte {
	if RECOMMENT.Match(src) {
		return src
	}
	re := REPARA.FindSubmatch(src)

	en := strings.TrimRight(string(re[2]), "\n")

	enstr := strings.ReplaceAll(en, "\n", " ")
	enstr = MultiSpace.ReplaceAllString(enstr, " ")
	enstr = strings.TrimSpace(enstr)

	if ja, ok := rep.catalog[enstr]; ok {
		para := fmt.Sprintf("$1<!--\n%s\n-->\n%s$3", en, strings.TrimRight(ja, "\n"))
		ret := REPARA.ReplaceAll(src, []byte(para))
		return ret
	}

	if rep.mt {
		fmt.Print("API問い合わせ...")
		ja := textraTranslate(enstr)
		ja = KUTEN.ReplaceAllString(ja, "。\n")
		para := fmt.Sprintf("$1<!--\n%s\n-->\n<!-- 機械翻訳 -->\n%s$3", en, strings.TrimRight(ja, "\n"))
		ret := REPARA.ReplaceAll(src, []byte(para))
		fmt.Print("Done\n")
		return ret
	}

	if rep.similar == 0 {
		return src
	}

	var maxdis float64
	den := ""
	dja := ""
	for dicen, dicja := range rep.catalog {
		distance := levenshtein.ComputeDistance(enstr, dicen)
		dis := (1 - (float64(distance) / float64(len(enstr)))) * 100
		if dis > maxdis {
			den = dicen
			dja = dicja
			maxdis = dis
		}
	}
	if maxdis > float64(rep.similar) {
		para := fmt.Sprintf("$1<!--\n%s\n-->\n<!-- マッチ度[%f]\n%s\n-->\n%s$3", en, maxdis, strings.TrimRight(den, "\n"), strings.TrimRight(dja, "\n"))
		ret := REPARA.ReplaceAll(src, []byte(para))
		return ret
	}
	return src
}

func replace(fileNames []string, mt bool, similar int) {
	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)
		rep := Rep{
			similar: similar,
			catalog: catalog,
			mt:      mt,
		}

		src, err := ReadFile(fileName)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			continue
		}

		ret := REPARA.ReplaceAllFunc(src, rep.Replace)

		if bytes.Equal(src, ret) {
			continue
		}

		fmt.Printf("replace: %s\n", fileName)
		out, err := os.Create(fileName)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			continue
		}
		fmt.Fprint(out, string(ret))
		out.Close()
	}
}

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "英語のパラグラフを「<!--英語-->日本語翻訳」に置き換える",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var mt bool
		var similar int
		var err error
		fmt.Printf("%#v\n", Conf.Name)
		if similar, err = cmd.PersistentFlags().GetInt("similar"); err != nil {
			log.Println(err)
			return
		}
		if mt, err = cmd.PersistentFlags().GetBool("mt"); err != nil {
			log.Println(err)
			return
		}
		if len(args) > 0 {
			replace(args, mt, similar)
			return
		}

		fileNames := targetFileName()
		replace(fileNames, mt, similar)
	},
}

func init() {
	rootCmd.AddCommand(replaceCmd)
	replaceCmd.PersistentFlags().IntP("similar", "s", 0, "Degree of similarity")
	replaceCmd.PersistentFlags().BoolP("mt", "", false, "Use machine translation")
}
