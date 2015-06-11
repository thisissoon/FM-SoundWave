# SOON\_ FM: Sound Wave

<img src="soundwave.jpg" width="90" height="100" align="right" />

Sound Wave is our Golang SOON\_ FM Spotify Player. It plays music and and publishes events
to a Redis Pub/Sub Service. It also listends for events on this service such as pause, resume
and stop events.

## OS Dependencies

SoundWave relies on 2 packages to be installed on your system:

* Portaudio 2 (`apt-get install portaudio19-dev` | `brew install portaudio`)
* libspotify: See https://pyspotify.mopidy.com/en/latest/installation/#installing-libspotify for
  install instructions.

## Install

Ensure you have Golang installed and your `$GOPATH` set correctly. Then run:

```
go get github.com/thisissoon/FM-Soundwave/...
go install github.com/thisissoon/FM-Soundwave/...
```

The binary will be installed to `$GOPATH/bin/soundwave`.
Update your path to include this: `export PATH=$PATH:$GOPATH/bin`

## Developing

To develope for soundwave you will need all the OS dependencies installed including Go. You will
also need `gpm`: https://github.com/pote/gpm

First make a directory that will be your `$GOPATH`, this is almost like a virtual environment for Go.

```
mkdir -p ~/.go/soundwave/src/github.com/thisissoon
```

Now set your `$GOPATH`:

```
export GOPATH=~/.go/soundwave/src/github.com/thisissoon
```

Now clone the repository to to a directory of your choice, for example:

```
mkdir -p ~/Development/Soundwave
cd ~/Development/Soundwave
git clone git@github.com:thisissoon/FM-SoundWave.git .
```

Now Symlink this directory into your `$GOPATH`, this is so Go can find the source:

```
ln -s ~/Development/Soundwave ~/.go/soundwave/src/github.com/thisissoon/FM-SoundWave
```

Next install the Depencides using GPM:

```
gpm install
```

You should now be able to build the Soundwave binary:

```
go build cmd/soundwave/soundwave.go
```

This will produce a ``soundwave`` binary at the root of the directory.

## Usage

SoundWave takes 7 arguments:

* `-u/--user`: Spotify User Name
* `-p/--pass`: Spotify Password
* `-k/--key`: Path to Spotify Key
* `-r/--redis`: Redis Server Address, defaults to `127.0.0.1:6379`
* `-c/--channel`: Redis Pub/Sub Channel Name
* `-q/--queue`: Redis Queue Key Name
* `-l/--log_level`: Log level - `info`, `debug`, `warn`, `error`, `fatal`, `panic`

```
soundwave -u foo -p -bar -k /spotify.key -c foo -q bar
```
