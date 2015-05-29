// Main Soundwave Package
//
// To build the soundwave binary simple run go build cmd/soundwave

package main

import (
	"fmt"

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
		fmt.Println(*(&spotify_track))
		player.Play(&spotify_user, &spotify_pass, &spotify_key, &spotify_track)
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
