package cmd

import (
	"log"

	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "辞書から英語と日本語訳のリストを出力する",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var filename, pre, enonly, jaonly, tsv bool
		var err error
		if filename, err = cmd.PersistentFlags().GetBool("filename"); err != nil {
			log.Println(err)
			return
		}
		if pre, err = cmd.PersistentFlags().GetBool("pre"); err != nil {
			log.Println(err)
			return
		}
		if enonly, err = cmd.PersistentFlags().GetBool("en"); err != nil {
			log.Println(err)
			return
		}
		if jaonly, err = cmd.PersistentFlags().GetBool("ja"); err != nil {
			log.Println(err)
			return
		}
		if tsv, err = cmd.PersistentFlags().GetBool("tsv"); err != nil {
			log.Println(err)
			return
		}

		if len(args) > 0 {
			if tsv {
				jpugdoc.TSVList(args)
				return
			}
			jpugdoc.List(filename, pre, enonly, jaonly, args)
			return
		}

		fileNames := targetFileName()
		if tsv {
			jpugdoc.TSVList(fileNames)
			return
		}
		jpugdoc.List(filename, pre, enonly, jaonly, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().BoolP("filename", "f", false, "filename")
	listCmd.PersistentFlags().BoolP("pre", "p", false, "prefix")
	listCmd.PersistentFlags().BoolP("en", "e", false, "English")
	listCmd.PersistentFlags().BoolP("ja", "j", false, "Japanese")
	listCmd.PersistentFlags().BoolP("tsv", "t", false, "tsv")
}
