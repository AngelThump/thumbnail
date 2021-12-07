package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	api "github.com/angelthump/thumbnail/api"
	utils "github.com/angelthump/thumbnail/utils"
)

var path string

func main() {
	cfgPath, err := utils.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	err = utils.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	path = filepath.FromSlash(utils.Config.Path)

	os.MkdirAll(path, 0755)

	var mainWg sync.WaitGroup
	mainWg.Add(1)
	check()
	mainWg.Wait()
}

func check() {
	streams := api.Find()
	if streams == nil {
		time.AfterFunc(60*time.Second, func() {
			check()
		})
		return
	}

	for _, stream := range streams {
		go saveThumbnail(stream)
	}

	time.AfterFunc(300*time.Second, func() {
		check()
	})
}

func saveThumbnail(stream api.Stream) {
	log.Printf("[%s] Executing ffmpeg: %s", stream.User.Username, "ffmpeg -y -i rtmp://"+stream.Ingest.Server+".angelthump.com/live/"+stream.User.Username+" -vframes 1 -f image2 "+path+stream.User.Username+".jpeg")
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", "rtmp://"+stream.Ingest.Server+".angelthump.com/live/"+stream.User.Username+"?key="+utils.Config.Ingest.AuthKey, "-vframes", "1", "-f", "image2", path+stream.User.Username+".jpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	log.Printf("[%s] Saved file at: %s", stream.User.Username, path+stream.User.Username+".jpeg")
}
