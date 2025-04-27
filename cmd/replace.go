package cmd

import (
	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "文書を「<!--英語-->日本語翻訳」に置き換える",
	Long: `抽出した辞書に基づいて文書を「<!--英語-->日本語翻訳」に置き換える。
文書により完全一致、類似文、機械翻訳で置き換える。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var update bool
		var mt bool
		var similar int
		var mts int
		var prompt bool
		var wip bool
		var err error
		if similar, err = cmd.PersistentFlags().GetInt("similar"); err != nil {
			return err
		}
		if update, err = cmd.PersistentFlags().GetBool("update"); err != nil {
			return err
		}
		if mts, err = cmd.PersistentFlags().GetInt("mts"); err != nil {
			return err
		}
		if mt, err = cmd.PersistentFlags().GetBool("mt"); err != nil {
			return err
		}
		if mt {
			mts = 90
		}
		if prompt, err = cmd.PersistentFlags().GetBool("prompt"); err != nil {
			return err
		}
		if wip, err = cmd.PersistentFlags().GetBool("wip"); err != nil {
			return err
		}

		fileNames, err := expandFileNames(args)
		if err != nil {
			return err
		}
		return jpugdoc.Replace(fileNames, vtag, update, similar, mts, wip, prompt)
	},
}

func init() {
	rootCmd.AddCommand(replaceCmd)
	replaceCmd.PersistentFlags().IntP("similar", "s", 0, "Degree of similarity")
	replaceCmd.PersistentFlags().BoolP("update", "u", false, "Update")
	replaceCmd.PersistentFlags().BoolP("mt", "", false, "Mark with machine translation")
	replaceCmd.PersistentFlags().IntP("mts", "", 0, "Use machine translation with similarity %")
	replaceCmd.PersistentFlags().BoolP("wip", "a", false, "Update even while work is in progress")
	replaceCmd.PersistentFlags().BoolP("prompt", "i", false, "Prompt before each replacement")
}
