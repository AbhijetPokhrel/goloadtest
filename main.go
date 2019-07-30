package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
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

	executor := loadExecutor()

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

func getTask(id string) *APIExecutorTask {
	task := APIExecutorTask{
		ID: fmt.Sprintf("%s%s", id, os.Getenv("CLIENT_ID")),
	}
	return &task
}

func loadExecutor() *Executor {
	executor := Executor{}
	executor.init()
	return &executor
}

func loadMasterModules(executor *Executor) {
	server := Server{
		executor: executor,
	}
	go executor.listen()
	server.init()
	srartGRPCServer(&server)
}

func loadSlaveModules(executor *Executor) {
	numOfReq, _ := strconv.Atoi(os.Getenv("NUM_OF_TASKS"))
	for i := 0; i < numOfReq; i++ {
		id := fmt.Sprintf("%d", i)
		executor.addTask(getTask(id))
	}
	setUpClient(executor)
}
