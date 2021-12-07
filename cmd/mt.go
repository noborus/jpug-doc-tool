package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/noborus/go-textra"
	"github.com/spf13/cobra"
)

// mtCmd represents the mt command
var mtCmd = &cobra.Command{
	Use:   "mt",
	Short: "APIを使用して文字列を翻訳する",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

func init() {
	rootCmd.AddCommand(mtCmd)
}
