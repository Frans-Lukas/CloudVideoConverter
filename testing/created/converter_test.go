package created

import (
	"errors"
	"math/rand"
	"testing"
	"time"
)
import converter "github.com/Frans-Lukas/cloudvideoconverter/server"

func TestTokenGenerateRandomString(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	str1 := converter.GenerateRandomString()
	str2 := converter.GenerateRandomString()
	if str1 == str2 {
		fatalFail(errors.New("random string generation is not random"))
	}

}
