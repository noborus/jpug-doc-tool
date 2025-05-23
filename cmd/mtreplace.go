package cmd

import (
	"fmt"

	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

// mtreplaceCmd represents the mtreplace command
var mtreplaceCmd = &cobra.Command{
	Use:   "mtreplace",
	Short: "機械翻訳のマークを実際に置き換える",
	RunE: func(cmd *cobra.Command, args []string) error {
		var limit int
		var err error
		if limit, err = cmd.PersistentFlags().GetInt("limit"); err != nil {
			return fmt.Errorf("failed to get 'limit' flag: %w", err)
		}
		fileNames, err := expandFileNames(args)
		if err != nil {
			return err
		}
		return jpugdoc.MTReplace(fileNames, limit, false)
	},
}

func init() {
	rootCmd.AddCommand(mtreplaceCmd)
	mtreplaceCmd.PersistentFlags().IntP("limit", "l", 400, "Limit for API queries")
}
