package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var activeWorkers int

func main() {
	println("Choose an option: ")
	for {
		println("1. Add a work creator.")
		println("2. Remove a work creator.")
		println("3. Create multiple VMs.")
		println("4. Kill all.")
		println("5. Exit.")
		action, err := readAction()
		if err != nil || action < 1 && action > 5 {
			println("Not a number!")
		} else {
			performAction(action)
		}
	}
}

func readAction() (int, error) {
	var action int
	_, err := fmt.Scanf("%d", &action)
	return action, err
}

func performAction(action int) {
	if action == 1 {
		addWorker()
	} else if action == 2 {
		removeWorker()
	} else if action == 3 {
		createMultipleWorkers()
	} else if action == 5 {
		os.Exit(0)
	} else if action == 4 {
		clearAll()
	}
}
func clearAll() {
	scriptPath := "./scripts/tfScripts/SpawningPool/killAllSpawningPools.sh"
	go func() {
		cmd := exec.Command(scriptPath)
		err := cmd.Run()
		if err != nil {
			log.Println("could kill all: " + err.Error())
		}
		println("Destroyed All VMs")
	}()
	println("Started kill all.")
	println()
	activeWorkers = 0
}

func createMultipleWorkers() {
	print("Number of VMs to create: ")
	numberOfVms, err := readAction()
	if err != nil || numberOfVms < 1 && numberOfVms > 20 {
		println("Not a number!")
	} else {
		scriptPath := "./scripts/tfScripts/SpawningPool/startSpawningPoolVM.sh"
		numberOfVms := strconv.Itoa(numberOfVms)
		go func() {
			cmd := exec.Command(scriptPath, numberOfVms)
			err := cmd.Run()
			if err != nil {
				log.Println("could not addWorker: " + err.Error())
			}
			println("Created new worker")
		}()
		activeWorkers, err = strconv.Atoi(numberOfVms)
		if err != nil {
			println("createMultipleWorkers failed VERY unexpectedly")
		}
		println("Created multiple vms thread")
	}

}

func addWorker() {
	scriptPath := "./scripts/tfScripts/SpawningPool/startSpawningPoolVM.sh"
	numberOfVms := strconv.Itoa(activeWorkers + 1)
	go func() {
		cmd := exec.Command(scriptPath, numberOfVms)
		err := cmd.Run()
		if err != nil {
			log.Println("could not addWorker: " + err.Error())
		}
		println("Created new worker")
	}()
	println("Started worker vm creation thread.")
	println()
	activeWorkers++
}

func removeWorker() {
	scriptPath := "./scripts/tfScripts/SpawningPool/killSpawningPoolVM.sh"
	go func() {
		cmd := exec.Command(scriptPath)
		err := cmd.Run()
		if err != nil {
			log.Println("could not addWorker: " + err.Error())
		}
		println("Destroyed VM 0")
	}()
	println("Started worker vm killing thread.")
	println()
	if activeWorkers > 0 {
		activeWorkers--
	}
}
