package video_converter

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

const url = "https://api.moogsoft.ai/express/v1/integrations/metrics"

func PrintKeyValue(key string, value int) {
	apiKey := os.Getenv("MOOGSOFT_API_KEY")
	if apiKey == "" {
		log.Printf("MOOGSOFT_API_KEY not set.")
		return
	}
	client := http.Client{}
	var jsonStr = []byte(`{"metric": "` + key + `", data: ` + strconv.Itoa(value) +
		`, "source": "www.videoconversionservice.com", "key": "dev", "tags": {"key": "value"}, "utc_offset": "GMT+01:00"}`)
	req, err := http.NewRequest("POST", url, jsonStr)
	if err != nil {
		println("Can't create request for ", url)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("apiKey-Type", apiKey)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		println("PrintKeyValue response failed! ", err.Error())
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("PrintKeyValue response failed! ", err.Error())
		return
	}
	println("response: ", body)
}