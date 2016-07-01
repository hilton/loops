package main
/*
Web service that plays audio loops, controlled by HTTP requests.
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

var players = make(map[string]*audio.Player)
var current *audio.Player

// Channel for controlling the player.
var transport = make(chan string)
var looping = make(chan bool)

// Plays a loop.
func play(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request.Method, request.URL)
	if (request.Method == "POST") {
		_, loop := request.URL.Query()["loop"]
		looping <- loop
		transport <- path.Base(request.URL.Path)
	}
	if (request.Method == "DELETE") {
		transport <- ""
	}
}

// Creates a player for the given file name.
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
		case name := <-transport:
			fmt.Printf("Play %s, loop = %v\n", name, loop)
			playing = players[name] != nil
			current = players[name]
		case newLoop := <-looping:
			loop = newLoop
		default:
			if (playing) {
				err := current.Play()
				if (err != nil) {
					panic(err)
				}
				for current.State() == audio.Playing {
				}
				if (!loop) {
					playing = false
				}
			}
		}
	}
}

// Sets up an audio player and the HTTP interface.
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
