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


To get started:
[x] modify mr/worker.go's Worker() to send RPC to coordinator asking for task
[x] modify coordinator to respond with file name of unstarted map task
	[x] create map of all files on start { filename: status }
	[x] create struct type with map and sync mutex
	[x] lock before accessing and modifying map
	[x] on request, send filename of file that is not started processing
3. modify worker to read the file
4. modify worker to call map function like in mrsequential.go
5. implement coordinator 10s wait for worker to complete else give to other worker
