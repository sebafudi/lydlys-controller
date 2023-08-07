package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"github.com/sebafudi/lydlys-controller/internal/config"
	"github.com/sebafudi/lydlys-controller/internal/connection"
)

type Save200 struct {
	Version    string        `json:"_version"`
	BPMChanges []interface{} `json:"_BPMChanges"`
	Events     []Event       `json:"_events"`
	Notes      []Note        `json:"_notes"`
	Obstacles  []interface{} `json:"_obstacles"`
	Bookmarks  []interface{} `json:"_bookmarks"`
}

type Save300 struct {
	Version                           string               `json:"version"`
	BPMEvents                         []interface{}        `json:"bpmEvents"`
	RotationEvents                    []interface{}        `json:"rotationEvents"`
	ColorNotes                        []interface{}        `json:"colorNotes"`
	BombNotes                         []interface{}        `json:"bombNotes"`
	Obstacles                         []interface{}        `json:"obstacles"`
	Sliders                           []interface{}        `json:"sliders"`
	BurstSliders                      []interface{}        `json:"burstSliders"`
	Waypoints                         []interface{}        `json:"waypoints"`
	BasicBeatmapEvents                []BasicBeatmapEvents `json:"basicBeatmapEvents"`
	ColorBoostBeatmapEvents           []interface{}        `json:"colorBoostBeatmapEvents"`
	LightColorEventBoxGroups          []interface{}        `json:"lightColorEventBoxGroups"`
	LightRotationEventBoxGroups       []interface{}        `json:"lightRotationEventBoxGroups"`
	BasicEventTypesWithKeywords       interface{}          `json:"basicEventTypesWithKeywords"`
	UseNormalEventsAsCompatibleEvents bool                 `json:"useNormalEventsAsCompatibleEvents"`
	CustomData                        interface{}          `json:"customData"`
}

type BasicBeatmapEvents struct {
	B  float64 `json:"b"`
	Et int     `json:"et"`
	I  int     `json:"i"`
	F  int     `json:"f"`
}

type Note struct {
	Time         float64 `json:"_time"` // in beats
	LineIndex    int     `json:"_lineIndex"`
	LineLayer    int     `json:"_lineLayer"`
	Type         int     `json:"_type"`
	CutDirection int     `json:"_cutDirection"`
}

type Event struct {
	Time  float64 `json:"_time"` // in beats
	Type  int     `json:"_type"`
	Value int     `json:"_value"`
}

func isLight(note Event) bool {
	return note.Type >= 0 && note.Type <= 4
}

func isLight300(event BasicBeatmapEvents) bool {
	return event.Et >= 0 && event.Et <= 4
}

func fadeLed(led [3]byte, strength byte) [3]byte {
	if led[0] < strength {
		led[0] = 0
	} else {
		led[0] -= strength
	}
	if led[1] < strength {
		led[1] = 0
	} else {
		led[1] -= strength
	}
	if led[2] < strength {
		led[2] = 0
	} else {
		led[2] -= strength
	}

	return led
}

func main() {
	err := config.ParseEnvs()
	if err != nil {
		fmt.Println(err)
		return
	}
	flags := config.GetFlags()
	connectionc := connection.StartConnection(*flags.Ip, *flags.Port)
	filePath := "tmp/paradise/Expert.json"

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var save Save200
	err = json.NewDecoder(file).Decode(&save)
	if err != nil {
		log.Fatal(err)
	}

	// const fps = 60
	const bpm = 70
	const ledOffset = 10
	const ledCount = 40
	beatTimeInMs := 60000.0 / bpm

	fileBytes, err := os.ReadFile("tmp/paradise/song.mp3")
	if err != nil {
		panic("reading my-file.mp3 failed: " + err.Error())
	}

	fileBytesReader := bytes.NewReader(fileBytes)
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}
	op := &oto.NewContextOptions{}
	op.SampleRate = decodedMp3.SampleRate()
	op.ChannelCount = 2
	op.Format = oto.FormatSignedInt16LE
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan
	player := otoCtx.NewPlayer(decodedMp3)

	notes := save.Events
	startTime := time.Now()
	player.Play()
	leds := make([][3]byte, 97)
	go func() {
		for {
			if !player.IsPlaying() {
				break
			}
			for i := 0; i < len(leds); i++ {
				leds[i] = fadeLed(leds[i], 1)
			}
			connection.SendUdpPacket(connectionc, leds)
			time.Sleep(1 * time.Millisecond)
		}
	}()
	for i := 0; i < len(notes); i++ {
		if !player.IsPlaying() {
			break
		}
		shouldContinue := make(chan bool, len(notes))
		noteTime := notes[i].Time
		go func() {
			for time.Since(startTime).Milliseconds() < int64(noteTime*beatTimeInMs) {
			}
			shouldContinue <- true
		}()
		<-shouldContinue
		if isLight(notes[i]) {
			// leds[notes[i].Et+ledOffset] = [3]byte{255, 255, 255}
			divideTo := 5
			if notes[i].Type == 0 {
				for j := 0; j < ledCount/divideTo; j++ {
					leds[j+ledOffset] = [3]byte{128, 0, 64}
				}
			} else if notes[i].Type == 1 {
				for j := ledCount / divideTo; j < ledCount/divideTo*2; j++ {
					leds[j+ledOffset] = [3]byte{128, 0, 64}
				}
			} else if notes[i].Type == 2 {
				for j := ledCount / divideTo * 2; j < ledCount/divideTo*3; j++ {
					leds[j+ledOffset] = [3]byte{128, 0, 64}
				}
			} else if notes[i].Type == 3 {
				for j := ledCount / divideTo * 3; j < ledCount/divideTo*4; j++ {
					leds[j+ledOffset] = [3]byte{128, 0, 64}
				}
			} else if notes[i].Type == 4 {
				for j := ledCount / divideTo * 4; j < ledCount; j++ {
					leds[j+ledOffset] = [3]byte{128, 0, 64}
				}
			}

		}

	}

}
