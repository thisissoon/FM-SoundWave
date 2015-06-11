// Main Soundwave Package
//
// To build the soundwave binary simple run go build cmd/soundwave

package main

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/thisissoon/FM-SoundWave"
	redis "gopkg.in/redis.v3"
	log "github.com/Sirupsen/logrus"
)

var (
	spotify_user  string
	spotify_pass  string
	spotify_key   string
	redis_address string
	redis_queue   string
	redis_channel string
	log_level     string
)

var soundWaveCmdLongDesc = `Sound Wave Plays Spotify Music for SOON_ FM`

var SoundWaveCmd = &cobra.Command{
	Use:   "soundwave",
	Short: "Play Spotify Music for SOON_ FM",
	Long:  soundWaveCmdLongDesc,
	Run: func(cmd *cobra.Command, args []string) {
		// Inicialze logger - set up given log level
		InitializeLogger()

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


// Sets up the logrus logger. Since this relies on command line flags it can
// only be setup as part of the persistent pre Run of the Myleene root command
func InitializeLogger() {
	switch log_level {
		case "info":
			log.SetLevel(log.InfoLevel)
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		case "fatal":
			log.SetLevel(log.FatalLevel)
		case "panic":
			log.SetLevel(log.PanicLevel)
		default:
			log.SetLevel(log.DebugLevel)
		}
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
	SoundWaveCmd.Flags().StringVarP(&log_level, "log_level", "l", "", "Log level")
}

func main() {
	SoundWaveCmd.Execute()
}
