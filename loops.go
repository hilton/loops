package main
/*
Web service that plays audio loops, controlled by HTTP requests.
*/

import (
	"fmt"
	"golang.org/x/mobile/exp/audio"
	"os"
	"net/http"
)

// Channel for controlling the player.
var transport = make(chan bool)

// Plays a loop.
// TODO Separate endpoints for play (once) and loop (indefinitely).
func play(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request.Method + " /play")
	transport <- true
}

// Stops playing a loop.
func stop(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request.Method + " /stop")
	transport <- false
}

// Creates a player for the given file name.
func player(fileName string) {
	// TODO Use standard resource path
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	player, err := audio.NewPlayer(file, 0, 0)
	defer player.Close()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Loaded %s\n", player)

	// TODO Fix audio click on loop.
	var playing = false
	for {
		select {
		case command := <-transport:
			playing = command
		default:
			if (playing) {
				err = player.Play()
				for player.State() == audio.Playing {
				}
			}
		}
	}
}

// Sets up an audio player and the HTTP interface.
func main() {
	// TODO Use command line arguments to register audio files to play.
	go player("beat.wav")

	http.HandleFunc("/play", play)
	http.HandleFunc("/stop", stop)
	http.ListenAndServe(":9000", nil)
}
