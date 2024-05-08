package cmd

import (
	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// wordCmd represents the word command
var wordCmd = &cobra.Command{
	Use:   "word",
	Short: "対応する単語が含まれているかチェックする",
	Long:  `英単語と日本語単語の対応が含まれているかチェックする`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var en, ja string
		if en, err = cmd.PersistentFlags().GetString("en"); err != nil {
			return err
		}
		if ja, err = cmd.PersistentFlags().GetString("ja"); err != nil {
			return err
		}

		fileNames := expandFileNames(args)
		return jpugdoc.CheckWord(en, ja, vtag, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(wordCmd)
	wordCmd.PersistentFlags().StringP("en", "e", "", "English")
	wordCmd.PersistentFlags().StringP("ja", "j", "", "Japanese")
}
