- Lab 1 Map Reduce
	- basic idea
		- one coordinator multiple workers on same machine
		- workers communicate with coordinator via RPC
		- workers
			- ask coordinator for task
			- read tasks input from one or more files
			- execute task
			- write task output to one or more files
			- if worker hasn't completed tasks in 10s give task to different worker
	- rules
		- map phase
			- divide intermediate keys into buckets for nReduce reduce tasks
			- each mapper creates nReduce intermediate files for reduce tasks
			- intermediate Map output files in current directory where workers can read them as input to reduce tasks
		- reduce phase
			- output of the X'th reduce tasks in the file mr-out-X
				- this file contains one line per reduce function output
				- watch out for correct format
		- main
			- expects mr/coordinator.go to implement Done() method to return true when MR job completed
			- when job done all workers should exit
	- hints

- Things to think about:
 	- The map part of your worker can use the ihash(key) function (in worker.go) to pick the reduce task for a given key.

- Run
	- cd src/main
	- go build -race -buildmode=plugin ../mrapps/wc.go
	- rm map-out*
	- go run -race mrcoordinator.go pg-*.txt
	- go build -race -buildmode=plugin ../mrapps/wc.go && go run -race mrworker.go wc.so


- Todo:
	- map
		- pass data into map function and store key value pairs in memory
		- write key value pairs to disk
		- pass location of intermediate files back to master
	- reduce
		- read all intermediate files
		- sort by intermediate keys
		- iterate over sorted intermediate data
		- pass each key and intermediate values to reduce function
	
	- change map worker
		- intermediate data structure is an array or arrays of KeyValue
		- use the hash function to get index
		- change storage 
	- Implement reduce functions
		- find way to convert data to valid form for reducef


