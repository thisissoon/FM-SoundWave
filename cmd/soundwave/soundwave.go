// Main Soundwave Package
//
// To build the soundwave binary simple run go build cmd/soundwave

package main

import (
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thisissoon/FM-SoundWave/events"
	"github.com/thisissoon/FM-SoundWave/perceptor"
	"github.com/thisissoon/FM-SoundWave/player"
)

var soundWaveCmdLongDesc = `Sound Wave Plays Spotify Music for SOON_ FM`

var SoundWaveCmd = &cobra.Command{
	Use:   "soundwave",
	Short: "Play Spotify Music for SOON_ FM",
	Long:  soundWaveCmdLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		// Set log level
		log.SetLevel(log.DebugLevel)

		// Make event channels
		channels := events.NewChannels()

		// Make Event Handler
		handler := events.NewHandler(channels)
		go handler.Run()

		// Create Perceptor
		pcptr := perceptor.New(
			viper.GetString("perceptor_address"),
			viper.GetString("secret"),
			handler.ReceiveChannel(),
			channels)
		go pcptr.WSConnection()

		// Create Player
		player, err := player.New(
			viper.GetString("spotify.user"),
			viper.GetString("spotify.pass"),
			viper.GetString("spotify.key"),
			pcptr,
			channels)
		if err != nil {
			// Exit on error
			log.Fatalf("Failed to create player: %s", err)
		}
		// Run the player - this will play the music
		go player.Run()

		// Channel to listen for OS Signals
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)

		// Run for ever unless we get a signal
		for sig := range signals {
			log.Println(sig)
			os.Exit(1)
		}
	},
}

func init() {
	// Load config from File
	log.SetLevel(log.WarnLevel)

	// Defaults
	viper.SetDefault("perceptor_address", "localhost:9000")
	viper.SetDefault("secret", "foo")
	viper.SetDefault("log_level", "warn")
	viper.SetDefault("spotify", map[string]string{
		"user": "CHANGE_ME",
		"pass": "CHANGE_ME",
		"key":  "CHANGE_ME",
	})

	// From file
	viper.SetConfigName("config")           // name of config file (without extension)
	viper.AddConfigPath("/etc/soundwave/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.soundwave") // call multiple times to add many search paths
	viper.AddConfigPath("$PWD/.soundwave")  // call multiple times to add many search paths
	err := viper.ReadInConfig()             // Find and read the config file
	if err != nil {                         // Handle errors reading the config file
		log.Warnf("No config file found or is not properly formatted: %s", err)
	}

	// Switch Log Level
	switch viper.GetString("log_level") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	SoundWaveCmd.Execute()
}
