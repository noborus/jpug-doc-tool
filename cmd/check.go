package cmd

import (
	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "文書をチェックする",
	Long:  `英語と日本語の文書から翻訳をチェックする`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cf, err := setCheckFlag(cmd)
		if err != nil {
			return err
		}
		fileNames, err := expandFileNames(args)
		if err != nil {
			return err
		}
		return jpugdoc.Check(fileNames, *cf)
	},
}

func setCheckFlag(cmd *cobra.Command) (*jpugdoc.CheckFlag, error) {
	cf := &jpugdoc.CheckFlag{
		VTag:     "",
		Ignore:   false,
		WIP:      false,
		Para:     false,
		Word:     false,
		Tag:      false,
		Num:      false,
		Author:   false,
		Sentence: false,
	}
	cf.VTag = vtag
	var err error
	if cf.Para, err = cmd.PersistentFlags().GetBool("para"); err != nil {
		return nil, err
	}
	if cf.Word, err = cmd.PersistentFlags().GetBool("word"); err != nil {
		return nil, err
	}
	if cf.Tag, err = cmd.PersistentFlags().GetBool("tag"); err != nil {
		return nil, err
	}
	if cf.Num, err = cmd.PersistentFlags().GetBool("num"); err != nil {
		return nil, err
	}
	if cf.Strict, err = cmd.PersistentFlags().GetBool("strict"); err != nil {
		return nil, err
	}
	if cf.Ignore, err = cmd.PersistentFlags().GetBool("ignore"); err != nil {
		return nil, err
	}
	if cf.WIP, err = cmd.PersistentFlags().GetBool("wip"); err != nil {
		return nil, err
	}
	if cf.Author, err = cmd.PersistentFlags().GetBool("author"); err != nil {
		return nil, err
	}
	if cf.Sentence, err = cmd.PersistentFlags().GetBool("sentence"); err != nil {
		return nil, err
	}
	return cf, nil
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.PersistentFlags().BoolP("para", "p", false, "para check")
	checkCmd.PersistentFlags().BoolP("word", "w", false, "Word check")
	checkCmd.PersistentFlags().BoolP("tag", "t", false, "Tag check")
	checkCmd.PersistentFlags().BoolP("num", "n", false, "Num check")
	checkCmd.PersistentFlags().BoolP("author", "x", false, "Author check")
	checkCmd.PersistentFlags().BoolP("strict", "s", false, "strict check")
	checkCmd.PersistentFlags().BoolP("ignore", "i", false, "Prompt before ignore registration")
	checkCmd.PersistentFlags().BoolP("wip", "a", false, "Work in progress check")
	checkCmd.PersistentFlags().BoolP("sentence", "e", false, "Sentence check")
}
