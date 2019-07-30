package goloadtest

import (
	"bufio"
	"context"
	"fmt"
	pb "goloadtest/result"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

const (
	// start to run the tasks
	startExec = "START"
	// stop running the tasks
	stopExec = "STOP"
	// pause runnnig the tasks
	pauseExec = "PAUSE"
	// resume if the executor is being paused
	resumeExec = "RESUME"
)

// Server is the grpc server for master
// only the master has the server runnning
// executor is the executor of the program
// clientStreams is the map of client Id to the client stream :: HERE client means the slaves
// totalSlaves is the total numner of slaves connected
// totalCompleted is the number of slaves who have completed the task
type Server struct {
	executor       *Executor
	clientStreams  map[string]pb.Result_ConntectMasterServer
	totalSlaves    int
	totalCompleted int
	// mutex          *sync.Mutex
}

func (server *Server) init() {
	server.clientStreams = make(map[string]pb.Result_ConntectMasterServer)
	server.totalSlaves = 0
	server.totalCompleted = 0
	// server.mutex = &sync.Mutex{}
}

// sendCommand send the command to the slaves having specific clientID
func (server *Server) sendCommand(clientID string, cmd *pb.ExecutionCommand) {
	stream := server.clientStreams[clientID]
	log.Printf("sending stream to client %s", clientID)
	stream.Send(cmd)
}

// ConntectMaster records the task result
// once the connection is made it will add the client strem to clientStreams map
// the strem is made to runnig forever so that we can send commands to client at specific event
func (server *Server) ConntectMaster(in *pb.ConnectionRequest, stream pb.Result_ConntectMasterServer) error {

	server.addNewClient(in.Id, stream)
	// make the stream long lived
	for {
	}

}

// PostSummary posts the summary to the master
func (server *Server) PostSummary(ctx context.Context, in *pb.Summary) (*pb.ExecutionCommand, error) {

	log.Printf("%s is last %t", in.Id, in.IsLast)
	server.executor.summary.successCount += int(in.SuccessCount)
	server.executor.summary.failedCount += int(in.FailedCount)
	// for i := 0; i < len(in.Results); i++ {

	// 	res := in.Results[i]
	// 	result := Result{
	// 		id:   res.Id,
	// 		msg:  res.Msg,
	// 		time: res.Time,
	// 	}
	// 	if res.IsError == true {
	// 		result.err = errors.New(res.Msg)
	// 	}
	// 	server.executor.resultChan <- result
	// 	// go server.executor.processResult(&result)
	// }
	// if in.IsLast == true {
	// 	server.NotifyTaskCompleted()
	// }

	cmd := pb.ExecutionCommand{
		Type: resumeExec,
	}

	if in.IsLast == true {
		// if its the last stream the master tells the slave to stop
		cmd.Type = stopExec
		// we will also record as one client completed its task
		server.NotifyTaskCompleted()
	}

	return &cmd, nil
}

// NotifyTaskCompleted is called when the task of a slave has been completed
// when all slave notifies about task completion the master will summarize the overall result
func (server *Server) NotifyTaskCompleted() {
	server.totalCompleted++
	if server.totalCompleted >= server.totalSlaves {
		log.Printf("All slaves completed their task")
		server.executor.generateSummary()
	}
}

// addNewClient add a client strem to clientStreams map based in clientId
func (server *Server) addNewClient(clientID string, stream pb.Result_ConntectMasterServer) {

	log.Printf("New Client : %s", clientID)
	if server.clientStreams[clientID] == nil {
		server.totalSlaves++
	}
	server.clientStreams[clientID] = stream

	log.Printf("registerd cliend %v", server.clientStreams[clientID])
}

// start Grpc serrver and wait for the start command to notify the client when to start
func srartGRPCServer(server *Server) {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// wait for the start command
	go waitForStart(server)
	options := []grpc.ServerOption{}
	options = append(options, grpc.MaxMsgSize(100*1024*1024))
	options = append(options, grpc.MaxRecvMsgSize(100*1024*1024))
	s := grpc.NewServer(options...)

	pb.RegisterResultServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

// waits fro string strart from the user so that it can notify connected slaves to start the tasks execution
func waitForStart(server *Server) {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	fmt.Println("IS start exec :: ", text)
	server.executor.summary.startTime = time.Now().UnixNano()
	if text == "start\n" {
		cmd := pb.ExecutionCommand{
			Type: startExec,
		}
		for clinetID := range server.clientStreams {
			server.sendCommand(clinetID, &cmd)
		}
	}
}
