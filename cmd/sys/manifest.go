package sys

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/ipc"
	"agent-telegram/internal/operations"
)

var manifestOpenAPI bool

// ManifestCmd outputs the machine-readable operation manifest.
var ManifestCmd = &cobra.Command{
	GroupID: "server",
	Use:     "manifest",
	Short:   "Output machine-readable operation manifest",
	Long: `Output the agentic operation manifest with input/output schemas, safety,
idempotency, retry hints, and confirmation requirements.

Use --openapi to output an OpenAPI 3.1 document for serve-api.`,
	Run: func(cmd *cobra.Command, _ []string) {
		var payload any
		if manifestOpenAPI {
			payload = operations.OpenAPI("agent-telegram API", "dev")
		} else {
			payload = map[string]any{
				"ok":         true,
				"operations": operations.Manifest(),
				"errorTypes": ipc.ErrorTypesManifest(),
			}
		}
		cliutil.NewRunnerFromCmd(cmd, true).PrintJSON(payload)
	},
}

// AddManifestCommand adds the manifest command to the root command.
func AddManifestCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ManifestCmd)
	ManifestCmd.Flags().BoolVar(&manifestOpenAPI, "openapi", false, "Output OpenAPI 3.1 document")
}
