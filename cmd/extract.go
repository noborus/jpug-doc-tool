package cmd

import (
	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "英語と日本語翻訳を抽出する",
	Long:  `jpug-docの文書から英語と日本語翻訳を抽出する`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileNames := expandFileNames(args)
		return jpugdoc.Extract(vtag, fileNames)
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
}
