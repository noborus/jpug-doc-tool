package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/noborus/jpug-doc-tool/jpugdoc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func targetFileName() []string {
	pattern := "./*.sgml"
	rePattern := "./*/*.sgml"

	fileNames, err := filepath.Glob(pattern)
	if err != nil {
		log.Println(err)
		return nil
	}
	reFileNames, err := filepath.Glob(rePattern)
	if err != nil {
		log.Println(err)
		return nil
	}
	fileNames = append(fileNames, reFileNames...)
	fileNames = jpugdoc.IgnoreFileNames(fileNames)
	return fileNames
}

func expandFileNames(args []string) []string {
	if len(args) == 0 {
		return targetFileName()
	}

	expandedArgs := []string{}
	for _, arg := range args {
		fileInfo, err := os.Stat(arg)
		if err != nil {
			log.Fatal(err)
		}
		if fileInfo.IsDir() {
			err := filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					expandedArgs = append(expandedArgs, path)
				}
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		} else {
			expandedArgs = append(expandedArgs, arg)
		}
	}
	expandedArgs = jpugdoc.IgnoreFileNames(expandedArgs)
	return expandedArgs
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "jpug-doc-tool",
	Version: jpugdoc.Version,
	Short:   "jpug-doc tool",
	Long: `
jpug-doc の翻訳を補助ツール。
前バージョンの翻訳を新しいバージョンに適用したり、
翻訳のチェックが可能です。`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(jpugdoc.InitJpug)

	_ = rootCmd.RegisterFlagCompletionFunc("", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jpug-doc-tool.yaml)")
	_ = rootCmd.RegisterFlagCompletionFunc("config", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.PersistentFlags().BoolVarP(&jpugdoc.Verbose, "verbose", "", false, "verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".jpug-doc-tool")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	if err := viper.Unmarshal(&jpugdoc.Config); err != nil {
		fmt.Println("config file Unmarshal error")
		fmt.Println(err)
		os.Exit(1)
	}
}
