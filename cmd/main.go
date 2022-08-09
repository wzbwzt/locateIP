package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "locateIP",
	Short: "locateIP",
}

var _ak string
var _platform string
var _filename string

func init() {
	rootCmd.PersistentFlags().StringVar(&_ak, "ak", "", "第三方key")
	rootCmd.PersistentFlags().StringVarP(&_platform, "plat", "p", "gaode",
		"第三方平台,枚举值eg.gaode/baidu 目前只支持gaode")
	rootCmd.PersistentFlags().StringVarP(&_filename, "file", "f", "", "解析的csv文件路径")

	// rootCmd.MarkPersistentFlagRequired("ak")

	rootCmd.AddCommand(NewLscmd())
	rootCmd.AddCommand(NewLocatecmd())
}

func Main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
