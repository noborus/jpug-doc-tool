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
			Para:   false,
			Word:   false,
			Tag:    false,
			Num:    false,
		}
		var err error
		var vTag string
		if vTag, err = cmd.PersistentFlags().GetString("vtag"); err != nil {
			log.Println(err)
			return
		}
		if cf.Para, err = cmd.PersistentFlags().GetBool("para"); err != nil {
			log.Println(err)
			return
		}
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
		if cf.Strict, err = cmd.PersistentFlags().GetBool("strict"); err != nil {
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

		jpugdoc.Check(fileNames, vTag, cf)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.PersistentFlags().StringP("vtag", "v", "", "original version tag")
	checkCmd.PersistentFlags().BoolP("para", "p", false, "para check")
	checkCmd.PersistentFlags().BoolP("word", "w", false, "Word check")
	checkCmd.PersistentFlags().BoolP("tag", "t", false, "Tag check")
	checkCmd.PersistentFlags().BoolP("num", "n", false, "Num check")
	checkCmd.PersistentFlags().BoolP("strict", "s", false, "strict check")
	checkCmd.PersistentFlags().BoolP("ignore", "i", false, "Prompt before ignore registration")
}
