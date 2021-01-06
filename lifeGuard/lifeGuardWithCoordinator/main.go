package main

import (
	"github.com/Frans-Lukas/cloudvideoconverter/lifeGuard/server"
	"os"
	"time"
)

func main() {
	coordinatorStatus := make(chan *bool)
	go server.StartLifeGuard(os.Args[1], os.Args[2], os.Args[3], coordinatorStatus)
	readCoordinatorStatus(coordinatorStatus)
}

func readCoordinatorStatus(in <- chan *bool) {
	
	for {
		select {
		case msg := <-in:
			if *msg {
				println("-------IS COORDINATOR-------")
			} else {
				println("-----IS NOT COORDINATOR-----")
			}
		default:

		}
		time.Sleep(time.Second)
	}
}