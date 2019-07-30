package client

// TODO ::
// test if the connection can be reconnected or not
// test if the success and failed counts are same in both server and masters as a whole

import (
	"fmt"
	"os"
	"testing"
)

func init() {

	os.Setenv("SERVER_IP", "localhost")
	os.Setenv("SERVER_IP", "localhost")
	os.Setenv("PORT", "5001")
	os.Setenv("CLIENT_ID", "clientTest")
	os.Setenv("NUM_OF_TASKS", "100")
	os.Setenv("NUM_OF_BATCH", "10")

	fmt.Printf("inint")
}

func TestClientConn(t *testing.T) {
	fmt.Printf("TestClietConn\n")
}

func TestClientSuccessCount(t *testing.T) {
	fmt.Printf("TestClietSuccess\n")
}

func TestClinetFailCounts(t *testing.T) {

}
