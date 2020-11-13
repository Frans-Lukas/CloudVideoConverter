package created

import (
	"errors"
	"github.com/Frans-Lukas/cloudvideoconverter/video_converter"
	"math/rand"
	"testing"
	"time"
)

func TestTokenGenerateRandomString(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	str1 := video_converter.GenerateRandomString()
	str2 := video_converter.GenerateRandomString()
	if str1 == str2 {
		fatalFail(errors.New("random string generation is not random"))
	}

}
