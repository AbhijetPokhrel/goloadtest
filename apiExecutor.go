package goloadtest

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// APIExecutorTask executes the api call
// This is a demonstration of a simple task
// Each task must implement the task interface.. see executor.go task interface
//
// init() - called when registering task to executor
// onSetUp() - is called before starting a task
// onStart() - is called when the task is executed..It is executed as go routine
// onEnd() - is called when the task is completed execution
//
// ID is the  id of the task
// resultChan is the channle to post the task result
// payload is the payuload in making api call
// err is the error occured during task execution
type APIExecutorTask struct {
	ID         string
	resultChan chan Result
	headers    map[string]string
	startTime  int64
	url        string
	err        error
}

// initialize the task
func (task *APIExecutorTask) init(resultChan chan Result) {
	task.resultChan = resultChan
}

func (task *APIExecutorTask) onSetup() {
	// fmt.Printf("\n%s onBegin", task.ID)
	task.startTime = time.Now().UnixNano()
	task.url = os.Getenv("GET_URL")
	task.headers = make(map[string]string)
}

func (task *APIExecutorTask) onStart() {

	request, err := http.NewRequest("GET", task.url, bytes.NewBuffer([]byte{}))
	if err != nil {
		task.err = err
		task.onEnd()
		return
	}

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		task.err = err
		task.onEnd()
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println(result)
	task.onEnd()
}

func (task *APIExecutorTask) onEnd() {
	// fmt.Printf("\n%s onEnd", task.ID)
	result := Result{
		id:   task.ID,
		msg:  "Complete",
		err:  task.err,
		time: (time.Now().UnixNano() - task.startTime) / 1e6,
	}
	// pubish the result to the channel
	task.resultChan <- result

}
