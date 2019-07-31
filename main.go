// Package goloadtest is the example of implementation of runnning test of huge number of task using compute
// power of different clusters(machines).
// Here grpc is used as a protocol for communication between the clusters.
// A cluster can be MASTER or SLAVE.
// MASTER cluster gives commands to the SLAVE clusters.
// SLAVE clusters will connect to MASTER cluster and listens for commands.
// We have used here APIExecutorTask as a demo task.
// This demo task does nothing but makes request to specific url as a get request as specified in .env file.
// We can run this task in multiple cluster to test how the url performs on heavy loads of requests.
// You can implemet you own version of the APIExecutorTask and make it performa specific task on multiple
// machines.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	// the mode in which the current program is running. It can be MASTER or SLAVE
	mode string
)

func main() {

	readArguments()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mode = os.Getenv("MODE")

	mode := os.Getenv("MODE")

	executor := getExcutor()

	if mode == "MASTER" {
		loadMasterModules(executor)
	} else {
		loadSlaveModules(executor)
	}
}

func readArguments() {
	args := os.Args
	os.Setenv("MODE", args[1])
	os.Setenv("SERVER_IP", args[2])
	os.Setenv("PORT", args[3])
	os.Setenv("CLIENT_ID", args[4])
	if os.Getenv("MODE") != "MASTER" {
		os.Setenv("NUM_OF_TASKS", args[5])
		os.Setenv("NUM_OF_BATCH", args[6])
	}
}

// returns a task interface pointer
// here APIExecutor is a task interface
// you can load your custom task here
func getTask(id string) *APIExecutorTask {
	task := APIExecutorTask{
		ID: fmt.Sprintf("%s%s", id, os.Getenv("CLIENT_ID")),
	}
	return &task
}

func getExcutor() *Executor {
	executor := Executor{}
	executor.init()
	return &executor
}

// loadMasterModules loads the neccesary params for master cluster
func loadMasterModules(executor *Executor) {
	server := Server{
		executor: executor,
	}
	go executor.listen()
	server.init()
	srartGRPCServer(&server)
}

// loadSlaveModules loads the prams necessary for slave cluster
func loadSlaveModules(executor *Executor) {
	numOfReq, _ := strconv.Atoi(os.Getenv("NUM_OF_TASKS"))
	for i := 0; i < numOfReq; i++ {
		id := fmt.Sprintf("%d", i)
		executor.addTask(getTask(id))
	}
	setUpClient(executor)
}
