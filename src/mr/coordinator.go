package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"sync"
)

type Coordinator struct {
	nReduce        int
	currentState   string
	workers        WorkersMap
	filesProcessed FilesProcessedMap
	// Your definitions here.
}

type WorkersMap struct {
	mu         sync.Mutex
	workersMap map[int]WorkerStatus
}

type WorkerStatus struct {
	status   string
	filename string
}

type FilesProcessedMap struct {
	mu      sync.Mutex
	filemap map[string]string
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//

func (c *Coordinator) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	workerId := args.WorkerId
	c.workers.mu.Lock()
	defer c.workers.mu.Unlock()
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

		c.workers.workersMap[workerId] = WorkerStatus{status: "processing", filename: filename}
		go checkIfWorkerIsStillRunning(workerId)

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
	c.filesProcessed.mu.Lock()
	defer c.filesProcessed.mu.Unlock()
	_, exists := c.filesProcessed.filemap[args.Filename]
	if !exists {
		fmt.Println(args.Filename + " was not found in filemap")
	}
	c.filesProcessed.filemap[args.Filename] = "processed"

}

func (c *Coordinator) NoticeMeSenpai(args *NoticeMeSenpaiArgs, reply *NoticeMeSenpaiReply) {
	c.workers.mu.Lock()
	defer c.workers.mu.Unlock()
	workerId := args.Id
	_, alreadyExists := c.workers.workersMap[workerId]

	if alreadyExists {
		reply.readyToWork = false
	} else {
		c.workers.workersMap[workerId] = WorkerStatus{status: "ready", filename: ""}
		reply.readyToWork = true
	}
}

func (c *Coordinator) AddFileNamesToMap() error {
	c.filesProcessed.filemap = map[string]string{}
	if c.currentState == "map" {
		for _, filename := range os.Args[2:] {
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			c.filesProcessed.filemap[filename] = "unprocessed"
			file.Close()
		}
	} else if c.currentState == "reduce" {
		files, err := os.ReadDir("./")
		if err != nil {
			log.Fatalf("error reading files in reduce state")
		}
		for _, file := range files {
			filename := file.Name()
			if strings.Contains(filename, "map-out-") {
				fileContents, err := os.Open(filename)
				if err != nil {
					log.Fatalf("cannot open %v", filename)
				}
				c.filesProcessed.filemap[filename] = "unprocessed"
				fileContents.Close()
			}
		}
	}

	return nil
}

func (c *Coordinator) FetchUnproccessedFileName() string {
	var unprocessedFilename string
	c.filesProcessed.mu.Lock()
	defer c.filesProcessed.mu.Unlock()
	for filename, status := range c.filesProcessed.filemap {
		if status == "unprocessed" {
			unprocessedFilename = filename
			c.filesProcessed.filemap[filename] = "processing"
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
	c.workers.workersMap = make(map[int]WorkerStatus)
	c.filesProcessed.filemap = make(map[string]string)
	c.currentState = "map"
	c.server()
	return &c
}
