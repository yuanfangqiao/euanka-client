// This example uses the PortAudio interface to stream the microphone thru Snowboy
// listening for the hotword.
//
// HOW TO USE:
// 	go run example/listen.go [path to snowboy resource file] [path to snowboy hotword file]
//
package main

import (
	"bytes"
	"encoding/binary"
	"euanka/snowboy"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
)

// Sound represents a sound stream with io.Reader interface.
type Sound struct {
	stream *portaudio.Stream
	data   []int16
}

// Read is the implementation of the io.Reader interface.
func (s *Sound) Read(p []byte) (int, error) {
	s.stream.Read()

	buf := &bytes.Buffer{}
	for _, v := range s.data {
		binary.Write(buf, binary.LittleEndian, v)
	}

	copy(p, buf.Bytes())
	return len(p), nil
}

func main() {
	inputChannels := 1
	outputChannels := 0
	sampleRate := 16000
	framesPerBuffer := make([]int16, 1024)

	// initialize the audio recording interface
	err := portaudio.Initialize()
	if err != nil {
		fmt.Errorf("Error initialize audio interface: %s", err)
		return
	}
	defer portaudio.Terminate()

	// open the sound input for the microphone
	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), len(framesPerBuffer), framesPerBuffer)
	if err != nil {
		fmt.Errorf("Error open default audio stream: %s", err)
		return
	}
	defer stream.Close()

	// open the snowboy detector
	d := snowboy.NewDetector(os.Args[1])
	defer d.Close()

	hub := snowboy.SetUpHub()

	d.HandleFunc(snowboy.NewHotword(os.Args[2], 0.5), func(name string) {
		fmt.Printf("Handle func for snowboy Hotword:%s \n", name)
		hub.GetListener().Focus = true
		hub.GetListener().StartSlienceCount = 0
	})

	d.HandleSilenceFunc(500*time.Millisecond, func(string) {
		fmt.Println("Silence 500 detected")
		// 监听中
		hub.GetListener().StartSlienceCount++
		if hub.GetListener().Focus && hub.GetListener().StartSlienceCount > 3 {
			hub.GetListener().Focus = false
			// 发送音频结束
			time.Sleep(100 * time.Millisecond)
			hub.SendWsClient("Done")
			log.Print("Focus end")
		}
	})

	sr, nc, bd := d.AudioFormat()
	fmt.Printf("sample rate=%d, num channels=%d, bit depth=%d\n", sr, nc, bd)

	err = stream.Start()
	if err != nil {
		fmt.Errorf("Error on stream start: %s", err)
		return
	}

	sound := &Sound{stream, framesPerBuffer}

	d.ReadAndDetect(sound, hub.GetListener())

}
