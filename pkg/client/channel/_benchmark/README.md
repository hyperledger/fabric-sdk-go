#Benchmark testing of Channel Client
    This benchmark test has 1 valid call of Channel Client's Execute() function
    
    Under the directory where this file resides, the test commands are run as shown under the below comments: 
	(
	    * on a Macbook Pro, warning messages are stripped out below for conciseness
	    * Benchmark is using Go's test command with -bench=ExecuteTx
	    * the -run=notest flag means execute a non-existant 'notest' in the current folder
	        This will avoid running normal unit tests along with the benchmarks
	    * by default, the benchmark tool decides when it collected enough information and stops
	    * the use of -benchtime=XXs forces the benchmark to keep executing until this time has elapsed
	        This allows the tool to run for longer times and collect more accurate information for larger execution loads
	    * the benchmark output format is as follows:
	        benchmarkname           [logs from benchamark tests-They have removed from the example commands below]   NbOfOperationExecutions     TimeInNanoSeconds/OperationExecuted   MemoryAllocated/OperationExecuted    NbOfAllocations/OperationExecuted  
	        Example from below commands:
	        BenchmarkExecuteTx-8    [logs removed]                                                                   100000                      164854 ns/op                          5743056 B/op                         50449 allocs/op 
	        
	    * the command output also shows the environment and the package used for the benchmark exection:
	        goos: darwin
            goarch: amd64
            pkg: github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/test/benchmark
	)

TODO Need a more controlled benchmark about channel client (perhaps analyze perf profiling data to investigate more fine grained memory/performance issues)

$ go test -run=notest -bench=ExecuteTx
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark
BenchmarkExecuteTx-8   	    1000	   1293015 ns/op	  219451 B/op	    3002 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark	2.294s
$ go test -run=notest -bench=ExecuteTx -benchtime=10s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark
BenchmarkExecuteTx-8   	   10000	   1308766 ns/op	  219258 B/op	    3001 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark	14.031s
$ go test -run=notest -bench=ExecuteTx -benchtime=30s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark
BenchmarkExecuteTx-8   	   30000	   1285455 ns/op	  219377 B/op	    3001 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark	52.563s
$ go test -run=notest -bench=ExecuteTx -benchtime=60s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark
BenchmarkExecuteTx-8   	  100000	   1373466 ns/op	  219192 B/op	    3001 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark	151.617s
$ go test -run=notest -bench=ExecuteTx -benchtime=120s
goos: darwin
goarch: amd64
pkg: github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark
BenchmarkExecuteTx-8   	  200000	   1347777 ns/op	  219278 B/op	    3001 allocs/op
PASS
ok  	github.com/hyperledger/fabric-sdk-go/pkg/client/channel/benchmark	284.201s


#Benchmark CPU & Memory performance analysis
    In order to generate profiling data for the chClient benchmark, the go test command can be extended to generate these.
    Note: If the below command complains about cpu.out or mem.out files are missing, create these files with empty content
     prior to running the command:
    
go test -v -run=notest -bench=ExecuteTx -benchtime=1s -outputdir ./bench1s -cpuprofile cpu.out -memprofilerate 1 -memprofile mem.out

    once ./bench1s has a valid cpu.out and mem.out content, then we can use go pprof command to examine the perf data.
    
    The above command will also generate the go binary file from which the profiling data is generated (benchmark.test).
    
    
    * For CPU perf data:
go tool pprof benchmark.test ./bench1s/cpu.out 

    * For Memory - allocation data (count by number of allocation):
go tool pprof --inuse_objects benchmark.test ./bench1s/mem.out 

    * For Memory - total allocation data:
go tool pprof --alloc_space benchmark.test ./bench1s/mem.out


    to generate the PDF report from the go tool pprof cli, simply execute
pdf
    
    or in svg format simply run:
svg