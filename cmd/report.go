package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/internal/config"
)

var CmdReport = &cobra.Command{
	Use:   "report [result_folder]",
	Short: "Generate HTML report from test results",
	Long: `Generate report.html from test results in the specified folder.
The folder should contain summary.json and optionally hrp.log files.

Examples:
  $ hrp report results/20250607234602/
  $ hrp report /path/to/test/results/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resultFolder := args[0]

		// Construct file paths
		summaryFile := filepath.Join(resultFolder, config.SummaryFileName)
		logFile := filepath.Join(resultFolder, config.LogFileName)
		reportFile := filepath.Join(resultFolder, config.ReportFileName)

		// Generate HTML report
		if err := hrp.GenerateHTMLReportFromFiles(summaryFile, logFile, reportFile); err != nil {
			return fmt.Errorf("failed to generate HTML report: %w", err)
		}

		log.Info().Str("report_file", reportFile).Msg("HTML report generated successfully")
		return nil
	},
}
