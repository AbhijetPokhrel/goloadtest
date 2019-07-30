package goloadtest

import (
	"context"
	"fmt"
	pb "goloadtest/result"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
)

var (
	// client is the current slave
	client pb.ResultClient
	// the stream form which the client will listen for master commands
	upStream *pb.Result_ConntectMasterClient
	// serverConn is the grpc connection to master
	serverConn *grpc.ClientConn
)

// connectToMaster connects the client(slave) to the master
// once the connection is made it will also wait form the server and listen for commands
func connectToMaster(executor *Executor) {

	stream, err := client.ConntectMaster(context.Background(), &pb.ConnectionRequest{
		Id: os.Getenv("CLIENT_ID"),
	})
	upStream = &stream
	if err != nil {
		log.Fatal(err)
	}

	// make stream long lived and wait for commands
	for {
		cmd, err := (*upStream).Recv()
		if err == io.EOF {
			// read done.
			log.Printf("Server closed the conn")
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a command : %v", err)
		}
		processServerCommand(executor, cmd)
	}

}

// postSummaryInBatch will post the summary results in certain interval of executor tasks list
// from is the task index of exectuor from where the batch summary is generated
// to is the task index of executor upto where the summary is to be generated
func postSummaryInBatch(executor *Executor, from int, to int) {

	summary := pb.Summary{
		SuccessCount: 0,
		FailedCount:  0,
		Results:      []*pb.TaskResult{},
		IsLast:       false,
	}

	for i := from; i < to; i++ {
		result := executor.summary.TaskResults[i]
		if result.err != nil {
			summary.FailedCount++
		} else {
			summary.SuccessCount++
		}
	}
	if to >= len(executor.summary.TaskResults) {
		summary.IsLast = true
	}

	summary.Id = fmt.Sprintf("%d-%d", from, to)
	// now lets post the summary to master
	postSummary(executor, &summary, 0)
}

// postSummary posts the summary result to the master
// retryCount is the number of times we will try to reconnect if the posting summary fails
func postSummary(executor *Executor, summary *pb.Summary, retryCount int) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd, err := client.PostSummary(ctx, summary)
	if err != nil {
		if retryCount < 3 {
			reconnect(executor)
			postSummary(executor, summary, retryCount+1)
		}
		log.Printf(err.Error())
		// log.Fatal("cannot send summary report")
	} else {
		processServerCommand(executor, cmd)
	}
}

// reconnect tries to reconnect to the serverF
func reconnect(executor *Executor) {
	serverConn.Close()
	setUpClient(executor)
}

// setUpClient connects the client to the master and makes necessary grpc dials
func setUpClient(executor *Executor) {

	var (
		address = fmt.Sprintf("%s:%s", os.Getenv("SERVER_IP"), os.Getenv("PORT"))
	)
	log.Printf("connection to : %s", address)

	options := []grpc.DialOption{}
	options = append(options, grpc.WithInsecure())
	options = append(options, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(100*1024*1024),
		grpc.MaxCallSendMsgSize(100*1024*1024)))
	conn, err := grpc.Dial(
		address,
		options...,
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	serverConn = conn
	client = pb.NewResultClient(conn)
	connectToMaster(executor)
}

// processServerCommand will process the server command form the master and performs task as specified
// see server.go for more details
func processServerCommand(executor *Executor, cmd *pb.ExecutionCommand) {

	if cmd.Type == startExec {
		// lets star the tasks
		executor.execute()
		return
	}

	if cmd.Type == stopExec {
		// the master has asked to stop
		log.Printf("%s", stopExec)
		executor.generateSummary()
		return
	}

	if cmd.Type == pauseExec {
		// TODO :: not implemented
		log.Printf("%s : No implemented yet ", pauseExec)
		return
	}

	if cmd.Type == resumeExec {
		// TODO :: not implemented
		log.Printf("%s : No implemented yet ", resumeExec)
		return
	}
}
