package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	locationFlags SendFlags
	locationLat   float64
	locationLong  float64
)

// LocationCmd represents the send location subcommand.
var LocationCmd = &cobra.Command{
	Use:   "location <peer>",
	Short: "Send a location",
	Long: `Send a geographic location to a Telegram user or chat.

Examples:
  agent-telegram send location @user --lat 40.7128 --long -74.0060
  agent-telegram send location --to @user --lat 51.5074 --long -0.1278`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = locationFlags.To.Set(args[0])
		}
		locationFlags.RequirePeer()

		runner := locationFlags.NewRunner()
		params := map[string]any{
			"latitude":  locationLat,
			"longitude": locationLong,
		}
		locationFlags.To.AddToParams(params)

		result := runner.CallWithParams("send_location", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_location")
		})
	},
}

func addLocationCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(LocationCmd)
	locationFlags.RegisterOptionalTo(LocationCmd)
	LocationCmd.Flags().Float64Var(&locationLat, "lat", 0, "Latitude")
	LocationCmd.Flags().Float64Var(&locationLong, "long", 0, "Longitude")
	_ = LocationCmd.MarkFlagRequired("lat")
	_ = LocationCmd.MarkFlagRequired("long")
}
