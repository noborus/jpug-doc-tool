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
	Long:  `抽出した辞書の英語と日本語訳のリストを出力する`,
	Run: func(cmd *cobra.Command, args []string) {
		var filename, affix, enOnly, jaOnly, tsv bool
		var err error
		if filename, err = cmd.PersistentFlags().GetBool("filename"); err != nil {
			log.Println(err)
			return
		}
		if affix, err = cmd.PersistentFlags().GetBool("pre"); err != nil {
			log.Println(err)
			return
		}
		if enOnly, err = cmd.PersistentFlags().GetBool("en"); err != nil {
			log.Println(err)
			return
		}
		if jaOnly, err = cmd.PersistentFlags().GetBool("ja"); err != nil {
			log.Println(err)
			return
		}
		if tsv, err = cmd.PersistentFlags().GetBool("tsv"); err != nil {
			log.Println(err)
			return
		}

		fileNames := expandFileNames(args)
		if tsv {
			jpugdoc.TSVList(fileNames)
			return
		}
		jpugdoc.List(filename, affix, enOnly, jaOnly, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().BoolP("filename", "f", false, "filename")
	listCmd.PersistentFlags().BoolP("pre", "p", false, "prefix/postfixも出力する")
	listCmd.PersistentFlags().BoolP("en", "e", false, "Englishのみ出力する")
	listCmd.PersistentFlags().BoolP("ja", "j", false, "Japaneseのみ出力する")
	listCmd.PersistentFlags().BoolP("tsv", "t", false, "tsv形式で出力する")
}
