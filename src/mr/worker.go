package mr

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
)

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// create worker id!
func getWorkerId() int {
	hash := fnv.New32()
	return int(hash.Sum32())
}

//

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	workerId := 0

	readyToWork := false

	for !readyToWork && workerId == 0 {
		workerId = getWorkerId()
		readyToWork = NoticeMeSenpai(workerId)
	}

Loop:
	for {
		status, filename, nReduce, err := RequestTask(workerId)
		if err != nil {
			fmt.Println(err)
			break
		}

		switch status {
		case "map":
			runMap(filename, mapf, nReduce)
		case "reduce":
			runReduce(filename, reducef)
		default:
			break Loop
		}

		ReportComplete(filename, workerId)
	}
}

func runMap(filename string, mapf func(string, string) []KeyValue, nReduce int) {
	contentsByte := getContentsOfFile(filename)
	contents := string(contentsByte)
	wordCounts := mapf(filename, contents)
	sliceLength := len(wordCounts) / nReduce

	sort.Sort(ByKey(wordCounts))
	for i := 0; i < nReduce; i += 1 {
		start := i * sliceLength
		end := start + sliceLength
		wordCountsSlice := wordCounts[start:end]
		writeWordCountsToFile(i, wordCountsSlice, filename)
	}
}

func runReduce(filename string, reducef func(string, []string) string) {
	contents := getContentsOfFile(filename)
	var keyValueData []string
	err := json.Unmarshal(contents, &keyValueData)

	if err != nil {
		fmt.Println("Error with json Unmarshal")
		return
	}

	data := make(map[string]int)

	for _, item := range keyValueData {
		count := reducef(item.Key, keyValueData)
	}

}

func getContentsOfFile(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	contentsBuffer, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()
	return contentsBuffer
}

func writeWordCountsToFile(count int, wordCountsSlice []KeyValue, filename string) {
	intermediateFileName := "map-out-" + filename + "-" + strconv.Itoa(count)
	intermediateFile, _ := os.Create(intermediateFileName)
	defer intermediateFile.Close()

	wordCountsJson, err := json.Marshal(wordCountsSlice)
	if err != nil {
		fmt.Println("error with converting slice to json")
	}

	_, err = intermediateFile.Write(wordCountsJson)

	if err != nil {
		fmt.Println("error writing to file")
	}
}

func RequestTask(workerId int) (string, string, int, error) {

	args := RequestTaskArgs{Status: "ready", WorkerId: workerId}
	reply := RequestTaskReply{}

	ok := call("Coordinator.RequestTask", &args, &reply)

	if !ok {
		fmt.Println("Error requesting task!")
		return "", "", 0, errors.New("error requesting task")
	}

	if reply.Filename == "" {
		return "", "", 0, errors.New("file name is null")
	}

	if reply.NReduce == 0 {
		return "", "", 0, errors.New("nReduce received is 0")
	}

	if reply.Status == "" {
		return "", "", 0, errors.New("status received is null")
	}

	return reply.Status, reply.Filename, reply.NReduce, nil
}

func ReportComplete(filename string, workerId int) {
	args := ReportCompleteArgs{Filename: filename, WorkerId: workerId}
	reply := ReportCompleteReply{}

	call("Coordinator.ReportComplete", &args, &reply)

}

func NoticeMeSenpai(workerId int) bool {
	args := NoticeMeSenpaiArgs{Id: workerId}
	reply := NoticeMeSenpaiReply{}

	ok := call("Coordinator.NoticeMeSenpai", &args, &reply)

	if !ok {
		fmt.Println("error trying to get senpai to notice me")
		return false
	}

	return reply.ReadyToWork
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
