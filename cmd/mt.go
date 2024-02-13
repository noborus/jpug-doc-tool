package cmd

import (
	"os"
	"strings"

	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// mtCmd represents the mt command
var mtCmd = &cobra.Command{
	Use:   "mt",
	Short: "APIを使用して文字列を翻訳する",
	Long:  `機械翻訳APIを使用して文字列を翻訳する`,
	RunE: func(cmd *cobra.Command, args []string) error {
		en := strings.Join(args, " ")
		en = strings.ReplaceAll(en, "\n", " ")
		w := os.Stdout
		return jpugdoc.MT(w, en)
	},
}

func init() {
	rootCmd.AddCommand(mtCmd)
}
