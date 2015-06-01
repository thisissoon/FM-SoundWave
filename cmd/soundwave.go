// Main Soundwave Package
//
// To build the soundwave binary simple run go build cmd/soundwave

package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/thisissoon/FM-SoundWave"
)

var (
	spotify_user  string
	spotify_pass  string
	spotify_key   string
	spotify_track string
	redis_uri     string
	redis_queue   string
)

var soundWaveCmdLongDesc = `Sound Wave Plays Spotify Music for SOON_ FM`

var SoundWaveCmd = &cobra.Command{
	Use:   "soundwave",
	Short: "Play Spotify Music for SOON_ FM",
	Long:  soundWaveCmdLongDesc,
	Run: func(cmd *cobra.Command, args []string) {

		s, a := soundwave.NewSession(&spotify_user, &spotify_pass, &spotify_key)

		defer s.Close() // Close Session
		defer a.Close() // Close Audio Writer

		// Play track in goroutine
		go func() {
			soundwave.Play(s, &spotify_track)
		}()

		// Log connection state changes
		go func() {
			for _ = range s.ConnectionStateUpdates() {
				var state string
				switch s.ConnectionState() {
				case 0:
					state = "Logged Out"
				case 1:
					state = "Logged In"
				case 2:
					state = "Disconnected"
				case 3:
					state = "Undefined / Unknown"
				case 4:
					state = "Offline"
				default:
					state = "Unknown State"
				}
				log.Println("Connection State:", state)
			}
		}()

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
	SoundWaveCmd.Flags().StringVarP(&spotify_user, "user", "u", "", "Spotify User")
	SoundWaveCmd.Flags().StringVarP(&spotify_pass, "pass", "p", "", "Spotify Password")
	SoundWaveCmd.Flags().StringVarP(&spotify_key, "key", "k", "", "Spotify Key Path")
	SoundWaveCmd.Flags().StringVarP(&spotify_track, "track", "t", "", "Spotify Track ID")
	SoundWaveCmd.Flags().StringVarP(&redis_uri, "redis", "r", "", "Redis Server URI")
	SoundWaveCmd.Flags().StringVarP(&redis_queue, "queue", "q", "", "Redis Queue Name")
}

func main() {
	SoundWaveCmd.Execute()
}
