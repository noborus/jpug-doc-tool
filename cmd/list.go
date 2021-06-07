package cmd

import (
	"fmt"
	"log"

	"github.com/jwalton/gchalk"
	"github.com/spf13/cobra"
)

func list(enonly bool, jaonly bool, fileNames []string) {
	for _, fileName := range fileNames {
		dicname := DICDIR + fileName + ".t"
		catalog := loadCatalog(dicname)
		for en, ja := range catalog {
			if !jaonly {
				fmt.Println(gchalk.Green(en))
			}
			if !enonly {
				fmt.Println(ja)
			}
			fmt.Println()
		}
	}
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "辞書から英語と日本語訳のリストを出力する",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var enonly, jaonly bool
		var err error
		if enonly, err = cmd.PersistentFlags().GetBool("en"); err != nil {
			log.Println(err)
			return
		}
		if jaonly, err = cmd.PersistentFlags().GetBool("ja"); err != nil {
			log.Println(err)
			return
		}

		if len(args) > 0 {
			list(enonly, jaonly, args)
			return
		}

		fileNames := targetFileName()
		list(enonly, jaonly, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().BoolP("en", "e", false, "English")
	listCmd.PersistentFlags().BoolP("ja", "j", false, "Japanese")
}
