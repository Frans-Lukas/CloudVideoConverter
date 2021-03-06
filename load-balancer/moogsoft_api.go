package video_converter

import (
	"bytes"
	"github.com/mitchellh/go-ps"
	"github.com/struCoder/pidusage"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const url = "https://api.moogsoft.ai/express/v1/integrations/metrics"

var apiKey = ""

func PrintKeyValue(key string, value int) {
	if apiKey == "" {
		readApiKey()
	}
	if apiKey == "" {
		log.Printf("MOOGSOFT_API_KEY not set.")
		return
	}
	println("using apikey: ", apiKey)
	client := http.Client{}
	var jsonStr = []byte(`{"metric": "` + key + `", "data": ` + strconv.Itoa(value) +
		`, "source": "www.videoconversionservice.com", "key": "dev", "tags": {"key": "value"}, "utc_offset": "GMT+01:00"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		println("Can't create request for ", url)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("apiKey", apiKey)
	println(string(jsonStr))
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
	println("response: ", string(body))
}
func readApiKey() {
	dat, err := ioutil.ReadFile("/tmp/MOOGSOFT_KEY.json")
	if err != nil {
		println("can't read moogsoft key file")
		return
	}
	apiKey = string(dat)
}

func PrintCPUUsage(vmName string) {
	processList, err := ps.Processes()
	if err != nil {
		log.Println("ps.Processes() Failed, are you using windows?")
		return
	}

	// map ages
	sum := 0.0
	for _, process := range processList {

		sysInfo, err := pidusage.GetStat(process.Pid())
		if err != nil {
			println("failed to get CPU usage..")
		} else {
			sum += sysInfo.CPU
		}
	}

	PrintKeyValue(vmName, int(sum))

}
