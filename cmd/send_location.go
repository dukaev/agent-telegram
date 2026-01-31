// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/telegram"
)

var (
	sendLocationFlags SendFlags
	sendLocationCity  string
)

func init() {
	sendLocationCmd := sendLocationFlags.NewCommand(CommandConfig{
		Use:   "send-location [latitude] [longitude]",
		Short: "Send a location to a Telegram peer",
		Long: `Send a location (geographical coordinates) to a Telegram user or chat.

You can specify coordinates directly or use --city flag to search by name.

Coordinates should be in decimal degrees:
  Latitude:  -90 to 90
  Longitude: -180 to 180

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
		Run: runSendLocation,
	})
	sendLocationCmd.Flags().StringVar(&sendLocationCity, "city", "", "Search by city/place name instead of coordinates")
	rootCmd.AddCommand(sendLocationCmd)
}

func runSendLocation(_ *cobra.Command, args []string) {
	runner := sendLocationFlags.NewRunner()

	var lat, lon float64
	var locationName string
	var err error

	if sendLocationCity != "" {
		lat, lon, locationName, err = geocodeCity(sendLocationCity)
	} else {
		lat, lon, err = parseCoordinates(args)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	params := map[string]any{
		"latitude":  lat,
		"longitude": lon,
	}
	sendLocationFlags.AddToParams(params)

	result := runner.CallWithParams("send_location", params)

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		lat, _ := rMap["latitude"].(float64)
		lon, _ := rMap["longitude"].(float64)
		fmt.Printf("Location sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		if locationName != "" {
			fmt.Printf("  Location: %s\n", locationName)
		}
		fmt.Printf("  Coordinates: %.6f, %.6f\n", lat, lon)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}

// geocodeCity converts a city name to coordinates.
func geocodeCity(city string) (lat, lon float64, locationName string, err error) {
	geocoder := telegram.NewGeoCoder()
	result, err := geocoder.Geocode(city)
	if err != nil {
		return 0, 0, "", err
	}
	fmt.Fprintf(os.Stderr, "Found: %s\n", result.DisplayName)
	return result.Lat, result.Lon, result.DisplayName, nil
}

// parseCoordinates parses coordinates from command line arguments.
func parseCoordinates(args []string) (lat, lon float64, err error) {
	if len(args) < 2 {
		return 0, 0, fmt.Errorf("latitude and longitude required (or use --city flag)")
	}

	_, err1 := fmt.Sscanf(args[0], "%f", &lat)
	_, err2 := fmt.Sscanf(args[1], "%f", &lon)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid coordinates. Use decimal degrees format")
	}

	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude must be between -180 and 180")
	}

	return lat, lon, nil
}
