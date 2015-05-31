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

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)

		go func() {
			for sig := range signals {
				fmt.Println(sig)
				os.Exit(1)
			}
		}()

		ids := []string{
			"7kFHcGLRFmUYajfLhaOcUK",
			"2FT9AEeUdZQsHaFu687AHy",
			"6poX3z4Zmx56OUfoXYkCAc",
		}

		s, a := player.NewSession(&spotify_user, &spotify_pass, &spotify_key)

		defer s.Close() // Close Session
		defer a.Close() // Close Audio Writer

		for _, id := range ids {
			fmt.Println(id)
			player.Play(s, &id)
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
