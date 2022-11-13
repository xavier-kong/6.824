package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

type Coordinator struct {
	nReduce      int
	currentState string
	// Your definitions here.

}

type filesProcessed struct {
	mu      sync.Mutex
	filemap map[string]string
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//

var filesProcessedMap = new(filesProcessed)

func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	if args.Status == "ready" {
		filename := c.FetchUnproccessedFileName()
		if filename == "" {
			if c.currentState == "map" {
				c.currentState = "reduce"
				c.AddFileNamesToMap()
				filename = c.FetchUnproccessedFileName()
			} else if c.currentState == "reduce" {
				c.currentState = "done"
			}
		}
		// add status here to reply
		*reply = RequestTaskReply{
			Status:   c.currentState,
			Filename: filename,
			NReduce:  c.nReduce,
		}
		return nil
	}
	return nil
}

func (c *Coordinator) ReportComplete(args *ReportCompleteArgs) {
	filesProcessedMap.mu.Lock()
	defer filesProcessedMap.mu.Unlock()
	_, exists := filesProcessedMap.filemap[args.Filename]
	if !exists {
		fmt.Println(args.Filename + " was not found in filemap")
	}
	filesProcessedMap.filemap[args.Filename] = "processed"

}

func (c *Coordinator) AddFileNamesToMap() error {
	filesProcessedMap.filemap = map[string]string{}
	if c.currentState == "map" {
		for _, filename := range os.Args[2:] {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			filesProcessedMap.filemap[filename] = "unprocessed"
			file.Close()
		}
	} else if c.currentState == "reduce" {
		// search  directory for all files with format map-out-{filename}-{count}
	}

	return nil
}

func (c *Coordinator) FetchUnproccessedFileName() string {
	var unprocessedFilename string
	filesProcessedMap.mu.Lock()
	defer filesProcessedMap.mu.Unlock()
	for filename, status := range filesProcessedMap.filemap {
		if status == "unprocessed" {
			unprocessedFilename = filename
			filesProcessedMap.filemap[filename] = "processing"
			break
		}
	}
	return unprocessedFilename
}

func (c *Coordinator) AddMapOutputToFileMap() {

}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.

	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{nReduce: nReduce}

	// Your code here.
	c.AddFileNamesToMap()
	c.currentState = "map"
	c.server()
	return &c
}
