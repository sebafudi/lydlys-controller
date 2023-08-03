package config

import (
	"flag"
	"os"
	"sync"
)

type Flags struct {
	Ip   *string
	Port *string
	FPS  *int
	Link *string
}

var flags Flags

var singleFlagLock = &sync.Mutex{}

type single struct {
}

var singleFlagLoad *single

func GetFlags() Flags {
	if singleFlagLoad == nil {
		singleFlagLock.Lock()
		defer singleFlagLock.Unlock()
		if singleFlagLoad == nil {
			singleFlagLoad = &single{}
			flags = parseFlags()
		}
	}
	return flags
}

func parseFlags() Flags {
	flags = Flags{
		Ip:   flag.String("ip", os.Getenv("DEFAULT_IP"), "IP address to send UDP packets to"),
		Port: flag.String("port", os.Getenv("DEFAULT_PORT"), "Port to send UDP packets to"),
		FPS:  flag.Int("fps", 60, "Frames per second"),
		Link: flag.String("link", "", "Link to download"),
	}
	flag.Parse()
	return flags
}
