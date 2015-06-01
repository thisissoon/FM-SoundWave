// Main Soundwave Package
//
// To build the soundwave binary simple run go build cmd/soundwave

package main

import (
	"fmt"
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
			fmt.Println(spotify_track)
			soundwave.Play(s, &spotify_track)
		}()

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)

		// Run for ever unless we get a signal
		for sig := range signals {
			fmt.Println(sig)
			os.Exit(1)
		}

	},
}

func init() {
	SoundWaveCmd.Flags().StringVarP(&spotify_user, "user", "u", "", "Spotify User")
	SoundWaveCmd.Flags().StringVarP(&spotify_pass, "pass", "p", "", "Spotify Password")
	SoundWaveCmd.Flags().StringVarP(&spotify_key, "key", "k", "", "Spotify Key Path")
	SoundWaveCmd.Flags().StringVarP(&spotify_track, "track", "t", "", "Spotify Track ID")
}

func main() {
	SoundWaveCmd.Execute()
}
