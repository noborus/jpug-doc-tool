package cmd

import (
	"log"

	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// wordCmd represents the word command
var wordCmd = &cobra.Command{
	Use:   "word",
	Short: "対応する単語が含まれているかチェックする",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var en, ja string
		if en, err = cmd.PersistentFlags().GetString("en"); err != nil {
			log.Println(err)
			return
		}
		if ja, err = cmd.PersistentFlags().GetString("ja"); err != nil {
			log.Println(err)
			return
		}
		if len(args) > 0 {
			jpugdoc.CheckWord(en, ja, args)
			return
		}

		fileNames := targetFileName()
		jpugdoc.CheckWord(en, ja, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(wordCmd)
	wordCmd.PersistentFlags().StringP("en", "e", "", "English")
	wordCmd.PersistentFlags().StringP("ja", "j", "", "Japanese")
}
