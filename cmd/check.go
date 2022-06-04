package cmd

import (
	"log"

	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "文書をチェックする",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cf := jpugdoc.CheckFlag{
			Ignore: false,
			Word:   false,
			Tag:    false,
			Num:    false,
		}
		var err error

		if cf.Word, err = cmd.PersistentFlags().GetBool("word"); err != nil {
			log.Println(err)
			return
		}
		if cf.Tag, err = cmd.PersistentFlags().GetBool("tag"); err != nil {
			log.Println(err)
			return
		}
		if cf.Num, err = cmd.PersistentFlags().GetBool("num"); err != nil {
			log.Println(err)
			return
		}
		if cf.Ignore, err = cmd.PersistentFlags().GetBool("ignore"); err != nil {
			log.Println(err)
			return
		}
		fileNames := targetFileName()
		if len(args) > 0 {
			fileNames = args
		}

		jpugdoc.Check(fileNames, cf)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.PersistentFlags().BoolP("word", "w", false, "Word check")
	checkCmd.PersistentFlags().BoolP("tag", "t", false, "Tag check")
	checkCmd.PersistentFlags().BoolP("num", "n", false, "Num check")
	checkCmd.PersistentFlags().BoolP("ignore", "i", false, "Prompt before ignore registration")
}
