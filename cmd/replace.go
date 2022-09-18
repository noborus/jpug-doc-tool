package cmd

import (
	"log"

	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "英語のパラグラフを「<!--英語-->日本語翻訳」に置き換える",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var vtag string
		var update bool
		var mt bool
		var similar int
		var prompt bool
		var err error
		if similar, err = cmd.PersistentFlags().GetInt("similar"); err != nil {
			log.Println(err)
			return
		}
		if update, err = cmd.PersistentFlags().GetBool("update"); err != nil {
			log.Println(err)
			return
		}
		if vtag, err = cmd.PersistentFlags().GetString("vtag"); err != nil {
			log.Println(err)
			return
		}
		if mt, err = cmd.PersistentFlags().GetBool("mt"); err != nil {
			log.Println(err)
			return
		}
		if prompt, err = cmd.PersistentFlags().GetBool("prompt"); err != nil {
			log.Println(err)
			return
		}

		if len(args) > 0 {
			jpugdoc.Replace(args, vtag, update, mt, similar, prompt)
			return
		}

		fileNames := targetFileName()
		jpugdoc.Replace(fileNames, vtag, update, mt, similar, prompt)
	},
}

func init() {
	rootCmd.AddCommand(replaceCmd)
	replaceCmd.PersistentFlags().IntP("similar", "s", 0, "Degree of similarity")
	replaceCmd.PersistentFlags().BoolP("update", "u", false, "Update")
	replaceCmd.PersistentFlags().StringP("vtag", "v", "", "original version tag")
	replaceCmd.PersistentFlags().BoolP("mt", "", false, "Use machine translation")
	replaceCmd.PersistentFlags().BoolP("prompt", "i", false, "Prompt before each replacement")
}
