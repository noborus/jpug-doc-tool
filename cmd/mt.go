package cmd

import (
	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// mtCmd represents the mt command
var mtCmd = &cobra.Command{
	Use:   "mt",
	Short: "APIを使用して文字列を翻訳する",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		jpugdoc.MT(args...)
	},
}

func init() {
	rootCmd.AddCommand(mtCmd)
}
