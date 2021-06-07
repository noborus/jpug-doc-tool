package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var DICDIR = "./.jpug-doc-tool/"

func ignoreFileNames(fileNames []string) []string {
	var ignoreFile map[string]bool = map[string]bool{
		"jpug-doc.sgml": true,
		"config0.sgml":  true,
		"config1.sgml":  true,
		"config2.sgml":  true,
		"config3.sgml":  true,
		"func0.sgml":    true,
		"func1.sgml":    true,
		"func2.sgml":    true,
		"func3.sgml":    true,
		"func4.sgml":    true,
	}

	ret := make([]string, 0, len(fileNames))
	for _, fileName := range fileNames {
		if ignoreFile[fileName] {
			continue
		}
		ret = append(ret, fileName)
	}
	return ret
}

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
	fileNames = ignoreFileNames(fileNames)
	return fileNames
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jpug-doc-tool",
	Short: "jpug-doc tool",
	Long: `
jpug-doc の翻訳を補助ツール。
前バージョンの翻訳を新しいバージョンに適用したり、
翻訳のチェックが可能です。`,
}

func ReadFile(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return src, nil
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
	cobra.OnInitialize(initJpug)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jpug-doc-tool.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initJpug() {
	f, err := filepath.Glob("./*.sgml")
	if err != nil || len(f) == 0 {
		fmt.Fprintln(os.Stderr, "*sgmlファイルがあるディレクトリで実行してください")
		fmt.Fprintln(os.Stderr, "cd github.com/pgsql-jp/jpug-doc/doc/src/sgml")
		return
	}
	if _, err := os.Stat(DICDIR); os.IsNotExist(err) {
		os.Mkdir(DICDIR, 0755)
	}
	refdir := DICDIR + "/ref"
	if _, err := os.Stat(refdir); os.IsNotExist(err) {
		os.Mkdir(refdir, 0755)
	}
}
