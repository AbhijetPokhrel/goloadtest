package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Executor executes the list of tasks
// The executor also waits for all the tasks to be completed
// Tasks is the  number of tasks waiting to be executed by the executor
// wg is the wait group to wait all the task execution*/
// resultChan is the channel to post task result
// isRunnung marks if the the executor is currently running
// summary is the summary of execution
type Executor struct {
	Tasks      []*Task
	wg         sync.WaitGroup
	resultChan chan Result
	isRunning  bool
	summary    Summary
	// mutex                       *sync.Mutex
}

// Summary is the summary of execution
// successCount is the number of success task executed
// failedCount is the number of failed task executed
// startTime is the time of start of the execution
// maxTimeTakingFailedTask is the failed task result taking max time
type Summary struct {
	successCount                int
	failedCount                 int
	startTime                   int64
	TaskResults                 []*Result
	maxTimeTakingSuccessfulTask *Result
	maxTimeTakingFailedTask     *Result
}

// Result is the result for the task execution
// msg is the message from the result
// err is the error occured
// id the result id
// time is the time taken to run the task
type Result struct {
	msg  string
	err  error
	id   string
	time int64
}

// Task is a operation to be performed
// evert task must implement these methods inorder to be executed
type Task interface {
	init(resultChan chan Result)
	onSetup()
	onStart()
	onEnd()
}

// init initialized executor params
func (executor *Executor) init() {
	executor.Tasks = []*Task{}
	executor.isRunning = false
	executor.resultChan = make(chan Result)
	executor.summary.successCount = 0
	executor.summary.failedCount = 0
	executor.summary.TaskResults = []*Result{}
	// executor.mutex = &sync.Mutex{}
}

// addTask adds a new task to the executor
func (executor *Executor) addTask(task Task) {
	task.init(executor.resultChan)
	executor.Tasks = append(executor.Tasks, &task)
}

// execute will execute the added tasks
func (executor *Executor) execute() {

	executor.summary.startTime = time.Now().UnixNano()
	// start to listen the task resuls
	go executor.listen()
	executor.runBatch()

}

// runBatch will the tasks in batch
// the batch number will be set to environment as NUM_OF_BATCH
// the batch number is added from cli argument in main.go
func (executor *Executor) runBatch() {

	tasksLen := len(executor.Tasks)
	log.Printf("Task lens :: %d", tasksLen)

	if tasksLen == 0 {
		executor.isRunning = false
		close(executor.resultChan)
		// notifyAllTaskCompleted(executor)
		// executor.generateSummary()
		return
	}

	batchSize, _ := strconv.Atoi(os.Getenv("NUM_OF_BATCH"))
	var to int

	if tasksLen >= batchSize {
		to = batchSize
	} else {
		to = tasksLen
	}
	// run all tasks
	for i := 0; i < to; i++ {
		executor.wg.Add(1)
		executor.startTask(i)
	}

	executor.Tasks = executor.Tasks[to:tasksLen]
	// wait for the task response
	executor.wg.Wait()
	// once one batch is compeleted run the next batch
	executor.runBatch()
}

// startTask will start a specific task form the index value
func (executor *Executor) startTask(index int) {
	task := executor.Tasks[index]
	(*task).onSetup()
	go (*task).onStart()
}

// listen will wait for all the result to be publishe in  the result channel
// no matter its server or client the resulst should be pulbished in the resultChan
func (executor *Executor) listen() {

	executor.isRunning = true
	total := 0
	from := 0
	to := 0
	// batch is the number of result to be sent in one summary
	// we are not sending the result execution details to master at once instead we send the summary of certain number
	// of task execution
	// here we are sending summary when 500 tasks are completed
	batch := 500

	for executor.isRunning == true {
		result := <-executor.resultChan
		if result.id == "" {
			// it marks that the channel has been closed
			break
		}
		total++
		executor.processResult(&result)
		if mode != "MASTER" {
			if (from + batch) <= len(executor.summary.TaskResults) {
				to = from + batch
			} else {
				to = len(executor.summary.TaskResults)
			}
			executor.summary.TaskResults = append(executor.summary.TaskResults, &result)
			// publishResult(&result)
			// log.Printf("total task %d gr : %d cpu : %d", total, runtime.NumGoroutine(), runtime.NumCPU())
			if len(executor.summary.TaskResults)%batch == 0 {
				// log.Printf("posting summary from %d to %d", from, to)
				go postSummaryInBatch(executor, from, to)
				from = to
			}
			executor.wg.Done()
		}
	}
	if mode != "MASTER" {
		// flush the summary if necessary
		postSummaryInBatch(executor, from, len(executor.summary.TaskResults))
	}
	log.Printf("stop listening now active go routinges %d", runtime.NumGoroutine())
}

func (executor *Executor) processResult(result *Result) {

	// executor.mutex.Lock()
	// defer executor.mutex.Unlock()

	if result.err != nil {
		//there was a error on task
		executor.processFailedResult(result)
	} else {
		// the task executed successfully
		executor.processSuccessResult(result)
	}
	// log.Printf("\ndone processing %s", result.id)
}

// handle the success result
func (executor *Executor) processSuccessResult(result *Result) {

	// fmt.Printf("Success task id : %s , msg : %s , time : %d\n", result.id, result.msg, result.time)
	executor.summary.successCount++
	if executor.summary.maxTimeTakingSuccessfulTask == nil {
		executor.summary.maxTimeTakingSuccessfulTask = result
		return
	}
	if executor.summary.maxTimeTakingSuccessfulTask.time < result.time {
		executor.summary.maxTimeTakingSuccessfulTask = result
	}
}

// handle the failed result
func (executor *Executor) processFailedResult(result *Result) {

	// fmt.Printf("* Err task id : %s , err : %s , time : %d\n", result.id, result.err, result.time)
	executor.summary.failedCount++
	if executor.summary.maxTimeTakingFailedTask == nil {
		executor.summary.maxTimeTakingFailedTask = result
		return
	}
	if executor.summary.maxTimeTakingFailedTask.time < result.time {
		executor.summary.maxTimeTakingFailedTask = result
	}
}

// generateSummary generates the summary of the execution
func (executor *Executor) generateSummary() {

	// stop listening
	// executor.isRunning = false
	if executor.summary.maxTimeTakingFailedTask == nil {
		executor.summary.maxTimeTakingFailedTask = &Result{
			id: "None",
		}
	}

	if executor.summary.maxTimeTakingSuccessfulTask == nil {
		executor.summary.maxTimeTakingSuccessfulTask = &Result{
			id: "None",
		}
	}

	fmt.Printf("-------------------------------------------------------------------\n")
	fmt.Printf("Summary\n")
	fmt.Printf("-------------------------------------------------------------------\n\n")

	fmt.Printf("Time taken                           :: %d(ms)\n", (time.Now().UnixNano()-executor.summary.startTime)/1e6)
	fmt.Printf("Successful tasks                     :: %d \n", executor.summary.successCount)
	fmt.Printf("Failed tasks                         :: %d \n", executor.summary.failedCount)
	fmt.Printf("Max time taking Successfull task     :: %s (%d)ms \n", executor.summary.maxTimeTakingSuccessfulTask.id, executor.summary.maxTimeTakingSuccessfulTask.time)
	fmt.Printf("Max time taking Failed task          :: %s (%d)ms \n", executor.summary.maxTimeTakingFailedTask.id, executor.summary.maxTimeTakingFailedTask.time)

	fmt.Printf("-------------------------------------------------------------------")

}
