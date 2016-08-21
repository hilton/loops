package main
/*
Web service that plays audio loops (a.k.a clips), controlled by HTTP requests.

Usage: start the server with a list of files
	loops [WAV files]

e.g. loops one.wav two.wav

Use the files’ base names (without .wav file extensions) as URL paths.

Queue an audio clip:
	http POST http://localhost:9000/one

Queue an audio clip to loop continuously:
	http POST 'http://localhost:9000/one?loop'

Stop playing at the end of the current clip:
	http DELETE 'http://localhost:9000/'
*/

import (
	"fmt"
	"golang.org/x/mobile/exp/audio"
	"os"
	"net/http"
	"path/filepath"
	"path"
	"strings"
)

var address = ":9000"

// Loaded audio players, and the current player.
var players = make(map[string]*audio.Player)
var current *audio.Player

// Command to play a named clip, with optional looping.
type command struct {
	clip string
	loop bool
}

// Channel for controlling the player.
var transport = make(chan command)

// Plays a clip.
func control(writer http.ResponseWriter, request *http.Request) {
	fmt.Printf("\n%s %s ", request.Method, request.URL)
	if (request.Method == "POST") {
		clip := path.Base(request.URL.Path)
		_, clipDefined := players[clip]
		if (clipDefined) {
			_, loop := request.URL.Query()["loop"]
			transport <- command{clip, loop}
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	} else if (request.Method == "DELETE") {
		fmt.Printf("\033[31m■\033[0m\n", )
		transport <- command{}
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Creates a player for the given file path.
func load(path string) (file *os.File, player *audio.Player) {
	clipName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	fmt.Printf("Loading %s\n", clipName)
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	player, err = audio.NewPlayer(file, 0, 0)
	if err != nil {
		panic(err)
	}

	players[clipName] = player
	return
}

// Controls the players, using the looping and transport channels.
func start() {
	var loop = false
	var playing = false
	for {
		for current.State() == audio.Playing {
		}

		select {
		case command := <-transport:
			playing = command.clip != ""
			current = players[command.clip]
			loop = command.loop
		default:
		}

		if playing {
			fmt.Printf("\033[32m▶ \033[0m", )
			err := current.Play()
			if (err != nil) {
				panic(err)
			}
			if !loop {
				playing  = false
			}
		}
	}
}

// Sets up audio players for the command line arguments and the HTTP interface.
func main() {
	for _, path := range os.Args[1:] {
		if (filepath.Ext(path) == ".wav") {
			file, player := load(path)
			defer file.Close()
			defer player.Close()
		}
	}
	go start()

	http.HandleFunc("/", control)
	fmt.Printf("Listening on address %s...\n", address)
	http.ListenAndServe(address, nil)
}
