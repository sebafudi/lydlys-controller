package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"io/fs"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"github.com/sebafudi/lydlys-controller/internal/config"
	"github.com/sebafudi/lydlys-controller/internal/connection"
	"github.com/sebafudi/lydlys-controller/internal/savefile"
	"github.com/wader/goutubedl"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func getId(link string) string {
	reg := regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?(?:youtube\.com|youtu\.be)\/(?:watch\?v=)?(.+)`)
	match := reg.FindStringSubmatch(link)
	if len(match) == 0 {
		return ""
	}
	return match[1]
}

func downloadYoutubeVideo(link string, mode string, output string) (string, float64, error) {
	goutubedl.Path = "yt-dlp"
	result, err := goutubedl.New(context.Background(), link, goutubedl.Options{})
	if err != nil {
		return "", 0, err
	}
	downloadResult, err := result.Download(context.Background(), mode)
	fmt.Println(result.Info.FPS)
	if err != nil {
		return "", 0, err
	}
	defer downloadResult.Close()

	id := getId(link)

	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		os.Mkdir("./tmp", 0755)
	}

	if _, err := os.Stat(fmt.Sprintf("./tmp/%s", id)); os.IsNotExist(err) {
		os.Mkdir(fmt.Sprintf("./tmp/%s", id), 0755)
	}

	f, err := os.Create(fmt.Sprintf("./tmp/%s/%s.%s", id, output, result.Info.Ext))
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	io.Copy(f, downloadResult)
	return result.Info.Ext, result.Info.FPS, nil
}

func convertAudioToMp3(path string, inputExt string) error {
	err := ffmpeg.Input(path + "/audio." + inputExt).Output(path + "/audio.mp3").Run()
	if err != nil {
		return err
	}
	return nil
}

func convertVideoToFrames(path string, inputExt string) error {
	if _, err := os.Stat(fmt.Sprintf("%s/frames", path)); os.IsNotExist(err) {
		os.Mkdir(fmt.Sprintf("%s/frames", path), 0755)
	}
	err := ffmpeg.Input(path+"/video."+inputExt).
		Output(path+"/frames/frame%03d.jpg", ffmpeg.KwArgs{"s": "97x2", "ss": "00:00:00"}).
		Run()
	return err
}

func checkForVideo(dir []fs.DirEntry) bool {
	_, err := findVideoInDir(dir)
	return err == nil
}

func findVideoInDir(dir []fs.DirEntry) (string, error) {
	for _, f := range dir {
		if f.Name()[:5] == "video" {
			return f.Name()[6:], nil
		}
	}
	return "", fmt.Errorf("no video file found")
}

func checkForAudio(dir []fs.DirEntry) bool {
	_, err := findAudioInDir(dir)
	return err == nil
}

func findAudioInDir(dir []fs.DirEntry) (string, error) {
	for _, f := range dir {
		if f.Name()[:5] == "audio" {
			return f.Name()[6:], nil
		}
	}
	return "", fmt.Errorf("no audio file found")
}

func main() {
	err := config.ParseEnvs()
	if err != nil {
		fmt.Println(err)
		return
	}
	flags := config.GetFlags()
	connectionc := connection.StartConnection(*flags.Ip, *flags.Port)
	link := "https://youtu.be/tXxFPYZWpA4"
	if *flags.Link != "" {
		link = *flags.Link
	}

	id := getId(link)
	path := fmt.Sprintf("./tmp/%s", id)

	var videoExt string
	var fps float64
	if dir, err := os.ReadDir(path); os.IsNotExist(err) || !checkForVideo(dir) {
		fmt.Println("Downloading video...")
		videoExt, fps, err = downloadYoutubeVideo(link, "bestvideo", "video")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		files, err := os.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}
		videoExt, err = findVideoInDir(files)
		if err != nil {
			log.Fatal(err)
		}
	}
	var audioExt string
	if dir, err := os.ReadDir(path); os.IsNotExist(err) || !checkForAudio(dir) {
		fmt.Println("Downloading audio...")
		audioExt, _, err = downloadYoutubeVideo(link, "bestaudio", "audio")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		files, err := os.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}
		audioExt, err = findAudioInDir(files)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/audio.mp3", path)); os.IsNotExist(err) {
		fmt.Printf("Converting audio to mp3...\n")
		err = convertAudioToMp3(path, audioExt)
		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/frames", path)); os.IsNotExist(err) {
		fmt.Println("Converting video to frames...")
		err = convertVideoToFrames(path, videoExt)
		if err != nil {
			log.Fatal(err)
		}
	}

	files, err := os.ReadDir(fmt.Sprintf("%s/frames", path))
	if err != nil {
		log.Fatal(err)
	}
	numberOfFrames := len(files)
	ledCount := 97

	leds := make([][][3]byte, numberOfFrames)
	for i := 0; i < numberOfFrames; i++ {
		leds[i] = make([][3]byte, ledCount)
	}

	for i := 1; i < numberOfFrames; i++ {
		file, err := os.Open(fmt.Sprintf("%s/frames/frame%03d.jpg", path, i))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		img, err := jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}

		for j := 0; j < ledCount; j++ {
			red, green, blue, _ := img.At(j, 0).RGBA()
			leds[i][j] = [3]byte{byte(red >> 8), byte(green >> 8), byte(blue >> 8)}
		}
	}
	savefile.SaveFile(leds, path+"/show.lyd", savefile.Metadata{FPS: fps, StripLength: 97})

	fileBytes, err := os.ReadFile(path + "/audio.mp3")
	if err != nil {
		panic("reading my-file.mp3 failed: " + err.Error())
	}

	fileBytesReader := bytes.NewReader(fileBytes)
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}
	op := &oto.NewContextOptions{}
	op.SampleRate = 48000
	op.ChannelCount = 2
	op.Format = oto.FormatSignedInt16LE
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan
	player := otoCtx.NewPlayer(decodedMp3)

	fmt.Println(fps)
	var frameDuration time.Duration = time.Second / time.Duration(fps)
	sinceStart := time.Now()
	lastFrame := 0
	avgDuration := time.Duration(10 * time.Second)
	skippedFrames := 0

	player.Play()
	for {
		if !player.IsPlaying() {
			break
		}
		start := time.Now()

		frameNumber := int(time.Since(sinceStart).Seconds() * float64(fps))
		if frameNumber >= numberOfFrames {
			fmt.Println("end of file")
			fmt.Printf("Duration: %v\n", time.Since(sinceStart))
			avgDuration = (avgDuration + time.Since(sinceStart)) / 2
			fmt.Printf("Avg Duration offset: %v\n", avgDuration-time.Duration(10*time.Second))
			fmt.Printf("Skipped frames: %v\n", skippedFrames)
			break
		}
		if frameNumber > lastFrame+1 {
			skippedFrames += frameNumber - lastFrame - 1
		}
		if frameNumber == lastFrame {
			continue
		}
		lastFrame = frameNumber
		connection.SendUdpPacket(connectionc, leds[frameNumber])
		for time.Since(start) < (frameDuration-time.Duration(time.Since(start).Milliseconds()))/2 {
		}

	}

}
