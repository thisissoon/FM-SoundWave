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
	redis "gopkg.in/redis.v3"
)

var (
	spotify_user  string
	spotify_pass  string
	spotify_key   string
	redis_address string
	redis_queue   string
	redis_channel string

	perceptorAddr string
	secret        string
)

var soundWaveCmdLongDesc = `Sound Wave Plays Spotify Music for SOON_ FM`

var SoundWaveCmd = &cobra.Command{
	Use:   "soundwave",
	Short: "Play Spotify Music for SOON_ FM",
	Long:  soundWaveCmdLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		// Create a new Player
		player, err := soundwave.NewPlayer(&spotify_user, &spotify_pass, &spotify_key)
		if err != nil {
			log.Fatalln(err) // Exit on failure to create player
		}

		defer player.Session.Close() // Close Session
		defer player.Audio.Close()   // Close Audio Writer

		// Create a redis client
		redis_client := redis.NewClient(&redis.Options{
			Network: "tcp",
			Addr:    redis_address,
		})

		// Create playlist
		playlist := soundwave.NewPlaylist(redis_queue, redis_channel, redis_client, player)
		// Watch playlist
		go playlist.Watch()

		// Publish track duration
		go playlist.CurrentTrackDurationPublisher()

		// Create Event Reactor
		reactor := soundwave.NewReactor(redis_channel, redis_client, player)
		// Spin off reactor consumer gorountiune
		go reactor.Consume()

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
	// Spotify Flags
	SoundWaveCmd.Flags().StringVarP(&spotify_user, "user", "u", "", "Spotify User")
	SoundWaveCmd.Flags().StringVarP(&spotify_pass, "pass", "p", "", "Spotify Password")
	SoundWaveCmd.Flags().StringVarP(&spotify_key, "key", "k", "", "Spotify Key Path")
	// Redis Flags
	SoundWaveCmd.Flags().StringVarP(&redis_address, "redis", "r", "127.0.0.1:6379", "Redis Server Address")
	SoundWaveCmd.Flags().StringVarP(&redis_queue, "queue", "q", "", "Redis Queue Name")
	SoundWaveCmd.Flags().StringVarP(&redis_channel, "channel", "c", "", "Redis Channel Name")
	// Perceptor Flags
	SoundWaveCmd.Flags().StringVarP(&perceptorAddr, "perceptor", "p", "perceptor.thisissoon.fm", "Perceptor Address")
	SoundWaveCmd.Flags().StringVarP(&perceptorAddr, "secret", "s", "CHANGE_ME", "Secret Key")
}

func main() {
	SoundWaveCmd.Execute()
}
