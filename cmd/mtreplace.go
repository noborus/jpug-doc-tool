package cmd

import (
	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// mtreplaceCmd represents the mtreplace command
var mtreplaceCmd = &cobra.Command{
	Use:   "mtreplace",
	Short: "機械翻訳のマークを実際に置き換える",
	RunE: func(cmd *cobra.Command, args []string) error {
		fileNames := expandFileNames(args)
		return jpugdoc.MTReplace(fileNames, false)
	},
}

func init() {
	rootCmd.AddCommand(mtreplaceCmd)
}
