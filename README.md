# SOON\_ FM: Sound Wave

<img src="soundwave.jpg" width="90" height="100" align="right" />

Sound Wave is our Golang SOON\_ FM Spotify Player. It plays music and and publishes events
to a Redis Pub/Sub Service. It also listends for events on this service such as pause, resume
and stop events.

# OS Dependencies

SoundWave relies on 2 packages to be installed on your system:

* Portaudio 2 (`apt-get install portaudio19-dev` | `brew install portaudio`)
* libspotify: See https://pyspotify.mopidy.com/en/latest/installation/#installing-libspotify for
  install instructions.

# Install

Ensure you have Golang installed and your `$GOPATH` set correctly. Then run:

```
go get github.com/thisissoon/FM-Soundwave/cmd
go build -o /usr/local/bin/soundwave github.com/thisissoon/FM-SoundWave/cmd
```

# Usage

SoundWave takes 6 arguments:

* `-u/--user`: Spotify User Name
* `-p/--pass`: Spotify Password
* `-k/--key`: Path to Spotify Key
* `-r/--redis`: Redis Server Address, defaults to `127.0.0.1:6379`
* `-c/--channel`: Redis Pub/Sub Channel Name
* `-q/--queue`: Redis Queue Key Name

```
soundwave -u foo -p -bar -k /spotify.key -c foo -q bar
```
