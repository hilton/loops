package main
/*
Web service that plays audio loops, controlled by HTTP requests.

Start the server with a list of files:
	loops [WAV files]

e.g. loops one.wav two.wav

Use the files’ base names (without .wav file extensions) as URL paths.

Queue audio:
	http POST http://localhost:9000/one

Queue audio to loop continuously:
	http POST 'http://localhost:9000/one?loop'

Stop playing at the end of the current loop:
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

// Listen address
var address = ":9000"

// Loaded audio players, and the current player.
var players = make(map[string]*audio.Player)
var current *audio.Player

// Channel for controlling the player.
var transport = make(chan string)
var looping = make(chan bool)

// Plays a loop.
// TODO Use a single channel, so the transport/looping command is consumed together.
// TODO Make a DELETE request asynchronous
func play(writer http.ResponseWriter, request *http.Request) {
	fmt.Printf("\n%s %s ", request.Method, request.URL)
	if (request.Method == "POST") {
		loopName := path.Base(request.URL.Path)
		_, loopDefined := players[loopName]
		if (loopDefined) {
			_, loop := request.URL.Query()["loop"]
			looping <- loop
			transport <- loopName
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	} else if (request.Method == "DELETE") {
		fmt.Printf("\033[31m■\033[0m\n", )
		transport <- ""
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Creates a player for the given file path.
func load(path string) (file *os.File, player *audio.Player) {
	loopName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	fmt.Printf("Loading %s\n", loopName)
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	player, err = audio.NewPlayer(file, 0, 0)
	if err != nil {
		panic(err)
	}

	players[loopName] = player
	return
}

// Controls the players, using the looping and transport channels.
func start() {
	// TODO Fix audio click on loop.
	var loop = false
	var playing = false
	for {
		select {
		case newLoop := <-looping:
			loop = newLoop
		case name := <-transport:
			playing = players[name] != nil
			current = players[name]
		default:
			if (playing) {
				fmt.Printf("\033[32m▶ \033[0m", )
				err := current.Play()
				if (err != nil) {
					panic(err)
				}
				for current.State() == audio.Playing {}
				if (!loop) {
					playing = false
				}
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

	http.HandleFunc("/", play)
	fmt.Printf("Listening on address %s...\n", address)
	http.ListenAndServe(address, nil)
}
