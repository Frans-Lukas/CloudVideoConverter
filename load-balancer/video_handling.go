package video_converter

import (
	"bytes"
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/constants"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func splitVideo(token string) {
	filePath := constants.LocalStorage + token + ".mp4"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf(errors.New("video to split does not exist").Error())
	}
	timeInSeconds, timeInSecondsString := getVideoTimeInSeconds(filePath)
	size := getVideoSize(filePath)
	numberOfSplits := int(math.Round(float64(size)/float64(sizeLimit) + 0.49))
	slizeSize := int(timeInSeconds) / numberOfSplits
	println("number of seconds: " + strconv.Itoa(int(timeInSeconds)))

	startTime := 0
	endTime := slizeSize
	for i := 1; i <= numberOfSplits; i++ {
		// must be string because of potential double inprecision
		println("start: ", startTime)
		println("end: ", endTime)
		performSplit(startTime, strconv.Itoa(endTime), filePath, i, numberOfSplits, token)

		startTime = endTime
		endTime += slizeSize
		if endTime > int(timeInSeconds)-slizeSize/2 {
			performSplit(startTime, timeInSecondsString, filePath, i+1, numberOfSplits, token)
			break
		}
	}

}

func mergeVideo(token string) error {
	videoParts, err := getVideoParts(token)

	if err != nil {
		return err
	}

	println("checking if parts are correctr")
	if correctVideoParts(videoParts) {
		println("parts are correct, merging")
		performMerge(videoParts, token)
		return nil
	}

	return errors.New("video parts are invalid")
}

func performMerge(videoParts []string, token string) {
	filePath := prepareFile(videoParts, token)

	format := extractFormat(videoParts[0])

	destinationFile := constants.LocalStorage + token + "AAAAB." + format
	command := "ffmpeg -f concat -safe 0 -i " + filePath + " -c copy " + destinationFile
	println(command)
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", filePath, "-c", "copy", destinationFile)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("could not pergormMerge: " + err.Error())
	}
}

func extractFormat(s string) string {
	format := strings.Split(s, ".")[1]
	return format
}

func prepareFile(videoParts []string, token string) string {
	filePath := constants.LocalStorage + token + "_parts.txt"

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Could not create file: " + err.Error())
	}
	defer file.Close()

	for _, v := range videoParts {
		file.WriteString("file '" + v + "'\n")
	}

	return filePath
}

func correctVideoParts(videoParts []string) bool {
	withoutExtension := strings.Split(videoParts[0], ".")[0]
	numberOfParts, err := strconv.Atoi(strings.Split(withoutExtension, "_")[1])
	if err != nil {
		println("can't find video part number: " + err.Error())
	}
	println("number of parts: " + strconv.Itoa(len(videoParts)))
	if len(videoParts) == numberOfParts {
		return true
	}
	return false
}
func getVideoParts(token string) ([]string, error) {
	listOfFiles := getFiles()
	parts := make([]string, 0)
	for _, v := range listOfFiles {
		if isAPart(token, v) {
			println("adding part " + v)
			parts = append(parts, v)
		}
	}

	if len(parts) == 0 {
		return parts, errors.New("split video parts not found")
	}
	return parts, nil
}

func isAPart(token string, potentialPart string) bool {
	matched, _ := regexp.MatchString(token+"-[0-9]+_[0-9]+", potentialPart)
	return matched
}

func getFiles() []string {
	files, err := ioutil.ReadDir(constants.LocalStorage)
	if err != nil {
		log.Fatalf("failed to get files:" + err.Error())
	}

	var fileList []string
	for _, f := range files {
		fileList = append(fileList, f.Name())
	}

	return fileList
}
func getVideoSize(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("invalid file " + filePath)
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		log.Fatalf("could not get file stats " + filePath)
	}
	return int(fi.Size())
}
func getVideoTimeInSeconds(filePath string) (float64, string) {
	println(filePath)
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf(errors.New("failed to get video time in seconds").Error())
	}
	timeString := out.String()
	timeString = strings.Replace(timeString, "\n", "", -1)
	timeInSec, err := strconv.ParseFloat(timeString, 64)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return timeInSec, timeString
}

func performSplit(startTime int, endTime string, filePath string, index int, total int, token string) error {
	targetFileName := constants.LocalStorage + token + "-" + strconv.Itoa(index) + "_" + strconv.Itoa(total) + ".mp4"
	str := "ffmpeg" + " -ss " + strconv.Itoa(startTime) + " -t " + endTime + " -i " + filePath + " -acodec " + " copy " + " -vcodec " + " copy " + targetFileName
	println(str)
	out, err := exec.Command("ffmpeg", "-ss", strconv.Itoa(startTime), "-t", endTime, "-i", filePath, "-acodec", "copy", "-vcodec", "copy", targetFileName).Output()
	println(string(out))
	if err != nil {
		println("failed to split video: " + err.Error() + ", file: " + filePath)
		return errors.New("failed to split video: " + filePath)
	}
	return nil
}
