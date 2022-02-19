package cmd

import (
	"log"

	"github.com/noborus/jpug-doc-tool/cmd/jpugdoc"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "文書をチェックする",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var word bool
		var tag bool
		var num bool
		var ignore bool
		var err error
		if word, err = cmd.PersistentFlags().GetBool("word"); err != nil {
			log.Println(err)
			return
		}
		if tag, err = cmd.PersistentFlags().GetBool("tag"); err != nil {
			log.Println(err)
			return
		}
		if num, err = cmd.PersistentFlags().GetBool("num"); err != nil {
			log.Println(err)
			return
		}
		if ignore, err = cmd.PersistentFlags().GetBool("ignore"); err != nil {
			log.Println(err)
			return
		}
		fileNames := targetFileName()
		if len(args) > 0 {
			fileNames = args
		}

		jpugdoc.Check(fileNames, ignore, word, tag, num)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.PersistentFlags().BoolP("word", "w", false, "Word check")
	checkCmd.PersistentFlags().BoolP("tag", "t", false, "Tag check")
	checkCmd.PersistentFlags().BoolP("num", "n", false, "Num check")
	checkCmd.PersistentFlags().BoolP("ignore", "i", false, "Prompt before ignore registration")
}
