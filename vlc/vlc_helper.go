package main

import (
	// Import standard libraries
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	//"time"
	//"io/ioutil"
	"path/filepath"

	// Import third party libraries
	log "github.com/Sirupsen/logrus" // Logger
	"github.com/jessevdk/go-flags"   // CMD Parser
)

type Opts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

	Folder string `short:"f" long:"folder" description:"Folder you would like to search through" value-name:"FOLDER" required:"true"`
}

func main() {
	opts := setup()
	//time.Sleep(.1 * time.Second)
	videoPath := getFilesInTree(opts.Folder)
	openWithVlc(videoPath)
}

func openWithVlc(path string) {
	cmd := exec.Command("vlc", "--fullscreen", path, "&")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("ERR:%s", err)
	}
}

// Recusivley searches a directory
// Returns first instance of media file with vlc compatible extension
func getFilesInTree(folder string) string {
	videoExts := [9]string{".mkv", ".mp4", ".mov", ".flv", ".avi", ".wmv", ".asf", ".wav", ".flac"}

	var videoPath = ""
	log.Trace("\nRootPath: %v", folder)
	err := filepath.Walk(folder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := strings.TrimSpace(filepath.Ext(path))
				log.Trace(fmt.Sprintf("\nPath:%s\nExt:%s", path, ext))
				for _, elem := range videoExts {
					if elem == strings.ToLower(ext) {
						videoPath = path
						return errors.New("File Found")
					}
				}
			}
			//fmt.Println(path, info.Size())
			return nil
		})
	if err == nil || videoPath == "" {
		panic("MEDIA FILE NOT FOUND WITHIN: " + folder)
	}
	log.Debug(videoPath)
	return videoPath
}

func setup() Opts {
	var opts Opts
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		panic(err)
	}
	log.SetOutput(os.Stdout)

	// Set Verbosity
	switch len(opts.Verbose) {
	case 1:
		log.SetLevel(log.WarnLevel)
	case 2:
		log.SetLevel(log.InfoLevel)
	case 3:
		log.SetLevel(log.DebugLevel)
	case 4:
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.FatalLevel)
	}

	if opts.Folder == "" {
		panic("FOLDER NOT SPECIFIED")
	}

	return opts
	/*
	   ----------Logger Examples---------
	   log.Trace("Something very low level.")
	   log.Debug("Useful debugging information.")
	   log.Info("Something noteworthy happened!")
	   log.Warn("You should probably take a look at this.")
	   log.Error("Something failed but I'm not quitting.")
	   // Calls os.Exit(1) after logging
	   log.Fatal("Bye.")
	   // Calls panic() after logging
	   log.Panic("I'm bailing.")
	*/
}
