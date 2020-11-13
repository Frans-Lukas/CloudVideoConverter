package created

import "log"

func failIfNotNull(err error, msg string) {
	if err != nil {
		log.Fatalf("error "+msg+": %v", err)
	}
}

func fatalFail(err error) {
	log.Fatalf("error: %v", err)
}
