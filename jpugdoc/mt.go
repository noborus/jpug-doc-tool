package jpugdoc

import (
	"fmt"
	"os"
	"strings"

	"github.com/noborus/go-textra"
)

func MT(args ...string) {
	c := Config
	config := textra.Config{}
	config.ClientID = c.ClientID
	config.ClientSecret = c.ClientSecret
	config.Name = c.Name
	cli, err := textra.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "textra: %s", err)
		os.Exit(1)
	}

	en := strings.Join(args, " ")
	en = strings.ReplaceAll(en, "\n", " ")
	ja, err := cli.Translate(c.APIAutoTranslateType, en)
	if err != nil {
		fmt.Fprintf(os.Stderr, "textra: %s", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s\n", c.APIAutoTranslateType, ja)

	jagen, err := cli.Translate(textra.GENERAL_EN_JA, en)
	if err != nil {
		fmt.Fprintf(os.Stderr, "textra: %s", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s\n", textra.GENERAL_EN_JA, jagen)
}
