package cmd

import (
	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
)

var conflictFormat string

// conflictCmd represents the conflict command
var conflictCmd = &cobra.Command{
	Use:   "conflict",
	Short: "同一英文で異なる日本語訳を出力する",
	Long:  `同じ英文に対して複数の日本語訳がある候補を出力する`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileNames, err := expandFileNames(args)
		if err != nil {
			return err
		}
		opt := jpugdoc.ConflictOption{Format: conflictFormat}
		return jpugdoc.Conflict(vtag, fileNames, opt)
	},
}

func init() {
	rootCmd.AddCommand(conflictCmd)
	conflictCmd.PersistentFlags().StringVar(&conflictFormat, "format", jpugdoc.ConflictFormatText, "output format: text|gh-md")
}
