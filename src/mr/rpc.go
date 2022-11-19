package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

type RequestTaskArgs struct {
	WorkerId int
	Status   string
}

type RequestTaskReply struct {
	Filename string
	NReduce  int
	Status   string
}

type ReportCompleteArgs struct {
	WorkerId int
	Filename string
}

type ReportCompleteReply struct {
}

type NoticeMeSenpaiArgs struct {
	Id int
}

type NoticeMeSenpaiReply struct {
	ReadyToWork bool
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
