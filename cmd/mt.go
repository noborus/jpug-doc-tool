package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// mtCmd represents the mt command
var mtCmd = &cobra.Command{
	Use:   "mt",
	Short: "APIを使用して文字列を翻訳する",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cl := apiClient(Config)
		en := strings.Join(args, " ")
		fmt.Println(cl.textraTranslate(en))
	},
}

func init() {
	rootCmd.AddCommand(mtCmd)
}
