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
	Long:  `英単語と日本語単語の対応が含まれているかチェックする`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var vTag string
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
			jpugdoc.CheckWord(en, ja, vTag, args)
			return
		}

		fileNames := targetFileName()
		jpugdoc.CheckWord(en, ja, vTag, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(wordCmd)
	wordCmd.PersistentFlags().StringP("vtag", "v", "", "original version tag")
	wordCmd.PersistentFlags().StringP("en", "e", "", "English")
	wordCmd.PersistentFlags().StringP("ja", "j", "", "Japanese")
}
