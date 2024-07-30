package api

import (
	"encoding/json"
	"log"

	utils "github.com/angelthump/thumbnail/utils"
	"github.com/go-resty/resty/v2"
)

type Stream struct {
	Created_at string `json:"createdAt"`
	Ingest     struct {
		Server string `json:"server"`
		Url    string `json:"url"`
	} `json:"ingest"`
	User struct {
		Username string `json:"username"`
	} `json:"user"`
}

func Find() []Stream {
	client := resty.New()
	resp, _ := client.R().
		SetHeader("X-Api-Key", utils.Config.StreamsAPI.AuthKey).
		Get(utils.Config.StreamsAPI.Hostname + "/streams")

	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		log.Printf("Unexpected status code, got %d %s", statusCode, string(resp.Body()))
		return nil
	}

	var streams []Stream
	err := json.Unmarshal(resp.Body(), &streams)
	if err != nil {
		log.Printf("Unmarshal Error %v", err)
		return nil
	}

	return streams
}
