package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

var activeWorkers int

func main() {
	println("Choose an option: ")
	for {
		action, err := readAction()
		if err != nil || action < 1 && action > 2 {
			println("Not a number!")
		} else {
			performAction(action)
		}
	}
}

func readAction() (int, error) {
	println("1. Add a work creator.")
	println("2. Remove a work creator.")
	var action int
	_, err := fmt.Scanf("%d", &action)
	return action, err
}

func performAction(action int) {
	if action == 1 {
		addWorker()
	} else if action == 2 {
		removeWorker()
	}
}

func addWorker() {
	scriptPath := "./scripts/tfScripts/SpawningPool/startSpawningPoolVM.sh"
	numberOfVms := strconv.Itoa(activeWorkers + 1)
	go func() {
		cmd := exec.Command(scriptPath, numberOfVms)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("could not addWorker: " + err.Error())
		}
		println("Created new worker")
	}()
	println("Started worker vm creation thread.")
	println()
	activeWorkers++
}

func removeWorker() {

}
