package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/httprunner/hrp/har2case"
)

// har2caseCmd represents the har2case command
var har2caseCmd = &cobra.Command{
	Use:   "har2case path...",
	Short: "Convert HAR to json/yaml testcase files",
	Long:  `Convert HAR to json/yaml testcase files`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("har2case called")
		var outputFiles []string
		for _, arg := range args {
			jsonPath, _ := har2case.NewHAR(arg).GenJSON()
			outputFiles = append(outputFiles, jsonPath)
		}
		log.Printf("%v", outputFiles)
	},
}

func init() {
	RootCmd.AddCommand(har2caseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// har2caseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// har2caseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
