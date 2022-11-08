package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
)

type Coordinator struct {
	nReduce int
	// Your definitions here.

}

// 0 = unprocessed
// 1 = processing
// 2 = processed
type filesProcessed struct {
	mu      sync.Mutex
	filemap map[string]int
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
		*reply = RequestTaskReply{
			Filename: filename,
			NReduce:  c.nReduce,
		}
		return nil
	}
	return nil
}

func (c *Coordinator) AddFileNamesToMap() error {
	for _, filename := range os.Args[2:] {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}
		filesProcessedMap.filemap[filename] = 0
		file.Close()
	}
	return nil
}

func (c *Coordinator) FetchUnproccessedFileName() string {
	var unprocessedFilename string
	filesProcessedMap.mu.Lock()
	defer filesProcessedMap.mu.Unlock()
	for filename, status := range filesProcessedMap.filemap {
		if status == 0 {
			unprocessedFilename = filename
			break
		}
	}
	return unprocessedFilename
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
	filesProcessedMap.filemap = make(map[string]int)
	c.AddFileNamesToMap()
	c.server()
	return &c
}
