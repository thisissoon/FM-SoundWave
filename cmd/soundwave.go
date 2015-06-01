// Main Soundwave Package
//
// To build the soundwave binary simple run go build cmd/soundwave

package main

import (
	"log"
	"os"
	"os/signal"
	"reflect"
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
	spotify_track string
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
		t := soundwave.LoadTrack(s, &spotify_track)

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
						v, err := client.LPop("lfoo").Result()
						if err == redis.Nil {
							// Key does not exist so no items on the queue, no need to log this, would be
							// very vebose
						} else if err != nil {
							log.Println(err)
						} else {
							log.Println(v)
							log.Println(reflect.TypeOf(v))
							soundwave.Play(s, p, t) // Blocks
						}
					}
				}
			}
		}()

		// Subscribes to a channel and listens for specific messages, pause, play etc
		go func() {
			log.Println("Subscribing to Channel:", redis_channel)

			pubsub := client.PubSub()
			pubsub.Subscribe(redis_channel)

			defer pubsub.Close()

			for {
				msgi, _ := pubsub.Receive()

				switch msg := msgi.(type) {
				case *redis.Subscription:
					log.Println(msg.Kind, msg.Channel)
				case *redis.Message:
					if msg.Payload == "pause" {
						p.Pause()
					}
					if msg.Payload == "play" {
						p.Play()
					}
					log.Println(msg.Channel, msg.Payload)
				default:
					log.Println("unknown message: %#v", msgi)
				}
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
	// Spotify Flags
	SoundWaveCmd.Flags().StringVarP(&spotify_user, "user", "u", "", "Spotify User")
	SoundWaveCmd.Flags().StringVarP(&spotify_pass, "pass", "p", "", "Spotify Password")
	SoundWaveCmd.Flags().StringVarP(&spotify_key, "key", "k", "", "Spotify Key Path")
	SoundWaveCmd.Flags().StringVarP(&spotify_track, "track", "t", "", "Spotify Track ID")
	// Redis Flags
	SoundWaveCmd.Flags().StringVarP(&redis_address, "redis", "r", "127.0.0.1:6379", "Redis Server Address")
	SoundWaveCmd.Flags().StringVarP(&redis_queue, "queue", "q", "", "Redis Queue Name")
	SoundWaveCmd.Flags().StringVarP(&redis_channel, "channel", "c", "", "Redis Channel Name")
}

func main() {
	SoundWaveCmd.Execute()
}
