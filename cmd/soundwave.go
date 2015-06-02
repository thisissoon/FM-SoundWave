// Main Soundwave Package
//
// To build the soundwave binary simple run go build cmd/soundwave

package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/op/go-libspotify/spotify"
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
)

var soundWaveCmdLongDesc = `Sound Wave Plays Spotify Music for SOON_ FM`

var SoundWaveCmd = &cobra.Command{
	Use:   "soundwave",
	Short: "Play Spotify Music for SOON_ FM",
	Long:  soundWaveCmdLongDesc,
	Run: func(cmd *cobra.Command, args []string) {

		s, a := soundwave.NewSession(&spotify_user, &spotify_pass, &spotify_key)
		p := s.Player()

		defer s.Close() // Close Session
		defer a.Close() // Close Audio Writer

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

		client := redis.NewClient(&redis.Options{
			Network: "tcp",
			Addr:    redis_address,
		})

		// Watches the current Queue, popping a track off the list and playing it
		go func() {
			for {
				tick := time.Tick(1 * time.Second)
				select {
				case <-tick:
					// We only want to play tracks if we are logged in, if we are not then
					// we will try again at the next tick
					if s.ConnectionState() == spotify.ConnectionStateLoggedIn {
						v, err := client.LPop(redis_queue).Result()
						if err == redis.Nil {
							// Key does not exist so no items on the queue, no need to log this, would be
							// very vebose
						} else if err != nil {
							log.Println(err)
						} else {
							t := soundwave.LoadTrack(s, &v)
							soundwave.Play(s, p, t) // Blocks
						}
					}
				}
			}
		}()

		// Event Reactor Routine
		go soundwave.EventReactor(&soundwave.ReactorConfig{
			RedisChannelName: redis_channel,
			RedisClient:      client,
			SpotifyPlayer:    p,
		})

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
}

func main() {
	SoundWaveCmd.Execute()
}
