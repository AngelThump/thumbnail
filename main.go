package main

import (
	b64 "encoding/base64"
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

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration, stream api.Stream) {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		saveThumbnail(stream)
		close(ch)
	}()
	select {
	case <-ch:
		return
	case <-time.After(timeout):
		log.Printf("[%s] Timed out saving thumbnail", stream.User.Username)
		return
	}
}

func check() {
	os.RemoveAll(path)
	os.MkdirAll(path, 0777)

	streams := api.Find()
	if streams == nil {
		time.AfterFunc(60*time.Second, func() {
			check()
		})
		return
	}

	maxGoroutines := 3
	guard := make(chan struct{}, maxGoroutines)

	for _, stream := range streams {
		guard <- struct{}{}
		go func(stream api.Stream) {
			var wg sync.WaitGroup
			WaitTimeout(&wg, 10*time.Second, stream)
			<-guard
		}(stream)
	}

	time.AfterFunc(300*time.Second, func() {
		check()
	})
}

func saveThumbnail(stream api.Stream) {
	base64String := b64.StdEncoding.EncodeToString([]byte(stream.Created_at + stream.User.Username))
	log.Printf("[%s] Executing ffmpeg: %s", stream.User.Username, "ffmpeg -y -i "+utils.Config.Ingest.Hostname+"/hls/"+base64String+"_"+stream.User.Username+"/index.m3u8 -vframes 1 -f image2 "+path+stream.User.Username+".jpeg")
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "panic", "-y", "-i", utils.Config.Ingest.Hostname+"/hls/"+base64String+"_"+stream.User.Username+"/index.m3u8", "-vframes", "1", "-f", "image2", path+stream.User.Username+".jpeg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	log.Printf("[%s] Saved file at: %s", stream.User.Username, path+stream.User.Username+".jpeg")
}
