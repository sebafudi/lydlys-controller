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

// const BackLasers = 0
// const RingLights = 1
// const LeftLasers = 2
// const RightLasers = 3
// const CenterLights = 4
// const BoostColors = 5
// const RingSpin = 8
// const RingZoom = 9
// const ExtraLights = 10
// const LeftLaserSpeed = 12
// const RightLaserSpeed = 13

const (
	BackLasers      = 0
	RingLights      = 1
	LeftLasers      = 2
	RightLasers     = 3
	CenterLights    = 4
	BoostColors     = 5
	RingSpin        = 8
	RingZoom        = 9
	ExtraLights     = 10
	LeftLaserSpeed  = 12
	RightLaserSpeed = 13
)

const (
	Off         = 0
	Blue        = 1
	FlashBlue   = 2
	FadeBlue    = 3
	FadeToBlue  = 4
	Red         = 5
	FlashRed    = 6
	FadeRed     = 7
	FadeToRed   = 8
	White       = 9
	FlashWhite  = 10
	FadeWhite   = 11
	FadeToWhite = 12
)

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
	Time       float64    `json:"_time"` // in beats
	Type       int        `json:"_type"`
	Value      int        `json:"_value"`
	CustomData CustomData `json:"_customData"`
}

type CustomData struct {
	Color         []float64     `json:"_color"`
	LightGradient LightGradient `json:"_lightGradient"`
}

type LightGradient struct {
	Duration   float64   `json:"_duration"`
	StartColor []float64 `json:"_startColor"`
	EndColor   []float64 `json:"_endColor"`
	Easing     string    `json:"_easing"`
}

func fadeColor(startColor [3]byte, endColor [3]byte, duration float64, time float64) [3]byte {
	var color [3]byte
	for i := 0; i < 3; i++ {
		color[i] = byte(float64(startColor[i]) + (float64(endColor[i])-float64(startColor[i]))*time/duration)

	}
	fmt.Println(time)
	return color
}

func isLight(note Event) bool {
	return note.Type >= 0 && note.Type <= 4
}

func isLight300(event BasicBeatmapEvents) bool {
	return event.Et >= 0 && event.Et <= 4
}

func processLed(led [3]byte, ledState byte, ledStateProgress *int, lastLedState byte, ledColor [3]byte) [3]byte {
	switch ledState {
	case Off:
		led = [3]byte{0, 0, 0}
	case Blue:
		if ledColor[0] > 0 || ledColor[1] > 0 || ledColor[2] > 0 {
			led = ledColor
		} else {
			led = [3]byte{220, 0, 0}
		}
	case FlashBlue:
		if *ledStateProgress == 0 {
			led = [3]byte{0, 0, 255}
		} else if *ledStateProgress < 2 {
			led = [3]byte{0, 0, 220}
		} else {
			if lastLedState == Off {
				led = [3]byte{0, 0, 0}
			} else if lastLedState == Blue {
				led = [3]byte{0, 0, 220}
			} else if lastLedState == Red {
				led = [3]byte{220, 0, 0}
			} else if lastLedState == White {
				led = [3]byte{220, 220, 220}
			} else {
				led = [3]byte{0, 0, 0}
			}
		}
	case FadeBlue:
		if *ledStateProgress == 0 {
			led = [3]byte{0, 0, 220}
		} else {
			if led[2] > 0 {
				led[2] -= 5
			}
		}
	case FadeToBlue:
		if led[0] > 0 {
			led[0] -= 5
		}
		if led[1] > 0 {
			led[1] -= 5
		}
		if led[2] < 255 {
			led[2] += 5
		}
	case Red:
		if ledColor[0] > 0 || ledColor[1] > 0 || ledColor[2] > 0 {
			led = ledColor
		} else {
			led = [3]byte{220, 0, 0}
		}
	case FlashRed:
		if *ledStateProgress == 0 {
			led = [3]byte{255, 0, 0}
		} else if *ledStateProgress < 2 {
			led = [3]byte{220, 0, 0}
		} else {
			if lastLedState == Off {
				led = [3]byte{0, 0, 0}
			} else if lastLedState == Blue {
				led = [3]byte{0, 0, 220}
			} else if lastLedState == Red {
				led = [3]byte{220, 0, 0}
			} else if lastLedState == White {
				led = [3]byte{220, 220, 220}
			} else {
				led = [3]byte{0, 0, 0}
			}

		}
	case FadeRed:
		if *ledStateProgress == 0 {
			led = [3]byte{220, 0, 0}
		} else {
			if led[0] > 0 {
				led[0] -= 5
			}
		}
	case FadeToRed:
		if led[0] < 255 {
			led[0] += 5
		}
		if led[1] > 0 {
			led[1] -= 5
		}
		if led[2] > 0 {
			led[2] -= 5
		}
	case White:
		led = [3]byte{220, 220, 220}
	case FlashWhite:
		if *ledStateProgress == 0 {
			led = [3]byte{255, 255, 255}
		} else if *ledStateProgress < 2 {
			led = [3]byte{220, 220, 220}
		} else {
			if lastLedState == Off {
				led = [3]byte{0, 0, 0}
			} else if lastLedState == Blue {
				led = [3]byte{0, 0, 220}
			} else if lastLedState == Red {
				led = [3]byte{220, 0, 0}
			} else if lastLedState == White {
				led = [3]byte{220, 220, 220}
			} else {
				led = [3]byte{0, 0, 0}
			}

		}
	case FadeWhite:
		if *ledStateProgress == 0 {
			led = [3]byte{220, 220, 220}
		} else {
			if led[0] > 0 {
				led[0] -= 5
			}
			if led[1] > 0 {
				led[1] -= 5
			}
			if led[2] > 0 {
				led[2] -= 5
			}
		}
	case FadeToWhite:
		if led[0] < 255 {
			led[0] += 5
		}
		if led[1] < 255 {
			led[1] += 5
		}
		if led[2] < 255 {
			led[2] += 5
		}
	}
	*ledStateProgress++
	return led
}

func processLeds(led [][3]byte, ledState []byte, ledStateProgress []int, lastLedState []byte, ledColor [][3]byte, ledStartColor [][3]byte, ledEndColor [][3]byte, ledFadeDuration []float64, beatsPerFrame float64) [][3]byte {
	for i := 0; i < len(led); i++ {
		if ledState[i] == 255 {
			led[i] = fadeColor(ledStartColor[i], ledEndColor[i], ledFadeDuration[i], float64(ledStateProgress[i])*beatsPerFrame)
			ledStateProgress[i]++
		} else {
			led[i] = processLed(led[i], ledState[i], &ledStateProgress[i], lastLedState[i], ledColor[i])
		}
	}
	return led
}

func float64ToByteColor(f []float64) [3]byte {
	if len(f) == 3 {
		return [3]byte{byte(f[0] * 255), byte(f[1] * 255), byte(f[2] * 255)}
	}
	if len(f) == 4 {
		return [3]byte{byte(f[0] * f[3] * 255), byte(f[1] * f[3] * 255), byte(f[2] * f[3] * 255)}
	}
	return [3]byte{0, 0, 0}
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
	fmt.Println(save.Events[0].CustomData)

	// const fps = 60
	const bpm = 172
	const ledOffset = 0
	const ledCount = 97
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
	ledState := make([]byte, 97)
	lastLedState := make([]byte, 97)
	ledStateProgress := make([]int, 97)
	ledColor := make([][3]byte, 97)
	ledStartColor := make([][3]byte, 97)
	ledEndColor := make([][3]byte, 97)
	ledFadeDuration := make([]float64, 97)
	fps := 120
	frameTime := time.Duration(1000/fps) * time.Millisecond
	beatsPerFrame := (bpm / 60) / float64(fps)
	fmt.Println(frameTime)
	go func() {
		for {
			start := time.Now()
			if !player.IsPlaying() {
				break
			}

			leds = processLeds(leds, ledState, ledStateProgress, lastLedState, ledColor, ledStartColor, ledEndColor, ledFadeDuration, beatsPerFrame)
			connection.SendUdpPacket(connectionc, leds)
			for time.Since(start) < time.Duration(1000/fps)*time.Millisecond {
			}
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
					if lastLedState[j+ledOffset] != ledState[j+ledOffset] {
						ledState[j+ledOffset] = lastLedState[j+ledOffset]
					}
					if notes[i].CustomData.LightGradient.Duration > 0 {
						ledState[j+ledOffset] = byte(255)
						ledStateProgress[j+ledOffset] = 0
						ledStartColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.StartColor)
						ledEndColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.EndColor)
						ledFadeDuration[j+ledOffset] = notes[i].CustomData.LightGradient.Duration
					} else {
						ledState[j+ledOffset] = byte(notes[i].Value)
						ledStateProgress[j+ledOffset] = 0
						ledColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.Color)
					}
				}
			} else if notes[i].Type == 1 {
				for j := ledCount / divideTo; j < ledCount/divideTo*2; j++ {
					if lastLedState[j+ledOffset] != ledState[j+ledOffset] {
						ledState[j+ledOffset] = lastLedState[j+ledOffset]
					}
					if notes[i].CustomData.LightGradient.Duration > 0 {
						ledState[j+ledOffset] = byte(255)
						ledStateProgress[j+ledOffset] = 0
						ledStartColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.StartColor)
						ledEndColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.EndColor)
						ledFadeDuration[j+ledOffset] = notes[i].CustomData.LightGradient.Duration
					} else {
						ledState[j+ledOffset] = byte(notes[i].Value)
						ledStateProgress[j+ledOffset] = 0
						ledColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.Color)
					}
				}
			} else if notes[i].Type == 2 {
				for j := ledCount / divideTo * 2; j < ledCount/divideTo*3; j++ {
					if lastLedState[j+ledOffset] != ledState[j+ledOffset] {
						ledState[j+ledOffset] = lastLedState[j+ledOffset]
					}

					if notes[i].CustomData.LightGradient.Duration > 0 {
						ledState[j+ledOffset] = byte(255)
						ledStateProgress[j+ledOffset] = 0
						ledStartColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.StartColor)
						ledEndColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.EndColor)
						ledFadeDuration[j+ledOffset] = notes[i].CustomData.LightGradient.Duration
					} else {
						ledState[j+ledOffset] = byte(notes[i].Value)
						ledStateProgress[j+ledOffset] = 0
						ledColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.Color)
					}
				}
			} else if notes[i].Type == 3 {
				for j := ledCount / divideTo * 3; j < ledCount/divideTo*4; j++ {
					if lastLedState[j+ledOffset] != ledState[j+ledOffset] {
						ledState[j+ledOffset] = lastLedState[j+ledOffset]
					}
					if notes[i].CustomData.LightGradient.Duration > 0 {
						ledState[j+ledOffset] = byte(255)
						ledStateProgress[j+ledOffset] = 0
						ledStartColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.StartColor)
						ledEndColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.EndColor)
						ledFadeDuration[j+ledOffset] = notes[i].CustomData.LightGradient.Duration
					} else {
						ledState[j+ledOffset] = byte(notes[i].Value)
						ledStateProgress[j+ledOffset] = 0
						ledColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.Color)
					}
				}
			} else if notes[i].Type == 4 {
				for j := ledCount / divideTo * 4; j < ledCount; j++ {
					if lastLedState[j+ledOffset] != ledState[j+ledOffset] {
						ledState[j+ledOffset] = lastLedState[j+ledOffset]
					}
					if notes[i].CustomData.LightGradient.Duration > 0 {
						ledState[j+ledOffset] = byte(255)
						ledStateProgress[j+ledOffset] = 0
						ledStartColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.StartColor)
						ledEndColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.LightGradient.EndColor)
						ledFadeDuration[j+ledOffset] = notes[i].CustomData.LightGradient.Duration
					} else {
						ledState[j+ledOffset] = byte(notes[i].Value)
						ledStateProgress[j+ledOffset] = 0
						ledColor[j+ledOffset] = float64ToByteColor(notes[i].CustomData.Color)
					}
				}
			}

		}

	}

}
