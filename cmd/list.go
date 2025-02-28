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
		var opt jpugdoc.ListOoptions
		var common bool
		var tsv bool
		var err error
		if common, err = cmd.PersistentFlags().GetBool("common"); err != nil {
			log.Println(err)
			return
		}
		if opt.WriteFile, err = cmd.PersistentFlags().GetBool("filename"); err != nil {
			log.Println(err)
			return
		}
		if opt.IsPre, err = cmd.PersistentFlags().GetBool("pre"); err != nil {
			log.Println(err)
			return
		}
		if opt.ENOnly, err = cmd.PersistentFlags().GetBool("en"); err != nil {
			log.Println(err)
			return
		}
		if opt.JAOnly, err = cmd.PersistentFlags().GetBool("ja"); err != nil {
			log.Println(err)
			return
		}
		if tsv, err = cmd.PersistentFlags().GetBool("tsv"); err != nil {
			log.Println(err)
			return
		}
		if opt.Strip, err = cmd.PersistentFlags().GetBool("strip"); err != nil {
			log.Println(err)
			return
		}
		if common {
			jpugdoc.ListCommon(opt)
			return
		}

		fileNames := expandFileNames(args)
		if tsv {
			jpugdoc.TSVList(true, fileNames)
			return
		}
		jpugdoc.List(opt, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().BoolP("filename", "f", false, "filename")
	listCmd.PersistentFlags().BoolP("common", "c", false, "common")
	listCmd.PersistentFlags().BoolP("pre", "p", false, "prefix")
	listCmd.PersistentFlags().BoolP("en", "e", false, "English")
	listCmd.PersistentFlags().BoolP("ja", "j", false, "Japanese")
	listCmd.PersistentFlags().BoolP("tsv", "t", false, "tsv")
	listCmd.PersistentFlags().BoolP("strip", "s", false, "strip")
}
