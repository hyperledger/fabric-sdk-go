#Benchmark testing of Channel Client
    This benchmark test has 1 valid call of Channel Client's Execute() function
    
    Under the directory where this file resides, the test commands are run as shown under the below comments: 
	(
	    * Benchmark is using Go's test command with -bench=Call* (or CallExecuteTx* or CallQuery*)
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
            pkg: github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel
            
        UPDATE: there are now 2 benchmarks in this file, the first being the original one which runs with sequential executions and the second makes use of
        parallel client executions to simulate simultanous calls from the client (command line outputs below are updated to reflect these two calls). 
            - To run the parallel test use -bench=ExecuteTxParallel
            - to run all benchmarks use -bench=* or -bench=ExecuteTx*
            
        NOTE: Since the peers/orderers are mocked and running in the same process as the benchmark's, the perf data for this benchmark includes info for both 
        the benchmark and the mocked servers which decreases the overall performance results. To get exact performance data for this channel client of the Go SDK, 
        one needs to run benchmarks against a real Fabric network with peers and orderers running in Docker containers.
        
        NOTE 2: The SDK config file must contain Fabric's perf configs in order to enable metrics collections. See this file for an example:
        test/fixtures/config/config_test.yaml
        
        NOTE 3: With the update to Fabric's metrics API, new metrics are automatically added when running the benchmark (without any additional setup).
        These are found in the statsd package and will show up in the Prometheus report as well: 
        internal/github.com/hyperledger/fabric/common/metrics/statsd/goruntime/collector.go 
        
        Final Note: Performance collection using the Metrics performance tool (see section below) now requires the SDK to be built with the pprof tag.
        This means in order to collect metrics data via the Prometheus report, the below sample commands were updated to include `-tags pprof`
	)

##### - Benchmark command examples:
    $ go test -tags pprof -run=notest -bench=Call*
    goos: darwin
    goarch: amd64
    pkg: github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel
    BenchmarkCallExecuteTx-8           	    1000	   1261314 ns/op	  208093 B/op	    3037 allocs/op
    BenchmarkCallQuery-8               	    2000	    975568 ns/op	  154871 B/op	    2429 allocs/op
    BenchmarkCallExecuteTxParallel-8   	    3000	    444365 ns/op	  208302 B/op	    3019 allocs/op
    BenchmarkCallQueryTxParallel-8     	    5000	    319143 ns/op	  154845 B/op	    2424 allocs/op
    PASS
    ok  	github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel	6.618s

    $ go test -tags pprof -run=notest -bench=Call* -benchtime=10s
    goos: darwin
    goarch: amd64
    pkg: github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel
    BenchmarkCallExecuteTx-8           	   10000	   1319398 ns/op	  208059 B/op	    3038 allocs/op
    BenchmarkCallQuery-8               	   10000	   1024998 ns/op	  154866 B/op	    2429 allocs/op
    BenchmarkCallExecuteTxParallel-8   	   30000	    564144 ns/op	  208060 B/op	    3015 allocs/op
    BenchmarkCallQueryTxParallel-8     	   50000	    386486 ns/op	  154849 B/op	    2424 allocs/op
    PASS
    ok  	github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel	68.788s
    
    $ go test -tags pprof -run=notest -bench=Call* -benchtime=30s
    goos: darwin
    goarch: amd64
    pkg: github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel
    BenchmarkCallExecuteTx-8           	   30000	   1295530 ns/op	  208024 B/op	    3038 allocs/op
    BenchmarkCallQuery-8               	   50000	   1027276 ns/op	  154858 B/op	    2429 allocs/op
    BenchmarkCallExecuteTxParallel-8   	  100000	    529863 ns/op	  208115 B/op	    3016 allocs/op
    BenchmarkCallQueryTxParallel-8     	  100000	    380491 ns/op	  154854 B/op	    2424 allocs/op
    PASS
    ok  	github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel	213.246s

    $ go test -tags pprof -run=notest -bench=Call* -benchtime=60s
    goos: darwin
    goarch: amd64
    pkg: github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel
    BenchmarkCallExecuteTx-8           	  100000	   1304653 ns/op	  208043 B/op	    3038 allocs/op
    BenchmarkCallQuery-8               	  100000	    988127 ns/op	  154850 B/op	    2429 allocs/op
    BenchmarkCallExecuteTxParallel-8   	  200000	    500477 ns/op	  208150 B/op	    3017 allocs/op
    BenchmarkCallQueryTxParallel-8     	  200000	    360922 ns/op	  154797 B/op	    2424 allocs/op
    PASS
    ok  	github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel	433.668s

    $ go test -tags pprof -run=notest -bench=Call* -benchtime=120s -timeout 30m
    goos: darwin
    goarch: amd64
    pkg: github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel
    BenchmarkCallExecuteTx-8           	  200000	   1286497 ns/op	  208050 B/op	    3038 allocs/op
    BenchmarkCallQuery-8               	  200000	    991448 ns/op	  154846 B/op	    2429 allocs/op
    BenchmarkCallExecuteTxParallel-8   	  500000	    510610 ns/op	  208140 B/op	    3017 allocs/op
    BenchmarkCallQueryTxParallel-8     	  500000	    365202 ns/op	  154773 B/op	    2424 allocs/op
    PASS
    ok  	github.com/hyperledger/fabric-sdk-go/test/performance/pkg/client/channel	926.270s
    
###Benchmark data (using Prometheus report)
    The Channel Client's Execute and Query functions have been amended to collect metric counts and time spent executing these functions.
    
    In order to support collecting the data, make sure to start the data collector server. An example of starting a data collector server is found 
    in this benchmark (reference chClient.StartOperationsSystem() call)
    
    then start the Prometheus Docker container (an example docker compose config file is found at:
    fabric-sdk-go/test/performance/prometheus
    )
    
    Finally run your sdk client and the perf data will be collected by the prometheus server. Navigate to 
    127.0.0.1:9095
    to view the report. 
    
    Make sure the Go client is running and some channel communication activity has occurred with a peer in order 
    to see collected performance data.
    
    
    For the purpose of this channel client benchmark, once the Prometheus docker container is started, run the benchmark with long enough
    run times and navigate to the address above to see data being collected 
    (run with -benchtime=300s will show this data on the report as an example)
    
    If you would like to collect perf data into your version of Prometheus server (example dedicated performance environment),
    make sure to create new metrics instances and register them the same way as in the channel client package.
    ie look at: "github.com/hyperledger/fabric-sdk-go/pkg/client/channel/chclientrun.go" to see how ClientMetrics is created and 
    metrics added in the code. 
    "github.com/hyperledger/fabric-sdk-go/pkg/client/channel/metrics.go" creates metrics structures to be used in the file above.
    
    currently, only channel client is configured with performance metrics (and operations system like Fabric).
    To setup data collection in your client application, see this file for more details: 
    fabric-sdk-go/pkg/client/channel/chclient.go
    
    The file fabric-sdk-go/test/performance/pkg/client/channel/chclient_fixture_test.go is loading metrics configs from the file referenced in `configPath` variable.

###Benchmark CPU & Memory performance analysis
    In order to generate profiling data for the chClient benchmark, the go test command can be extended to generate these.
    Note: If the below command complains about cpu.out or mem.out files are missing, create these files with empty content
     prior to running the command:
    
    $ go test -v -tags pprof -run=notest -bench=Call -benchtime=1s -outputdir ./bench1s -cpuprofile cpu.out -memprofilerate 1 -memprofile mem.out

    once ./bench1s has a valid cpu.out and mem.out content, then we can use go pprof command to examine the perf data.
    
    The above command will also generate the go binary file from which the profiling data is generated (benchmark.test).
    
    
    * For CPU perf data analysis:
    $ go tool pprof benchmark.test ./bench1s/cpu.out 

    * For Memory - allocation data analysis (count by number of allocation):
    $ go tool pprof --inuse_objects benchmark.test ./bench1s/mem.out 

    * For Memory - total allocation data analysis:
    $ go tool pprof --alloc_space benchmark.test ./bench1s/mem.out


    to generate the PDF report from the go tool pprof cli, simply execute:
    $ pdf
    
    or in svg format simply run:
    $ svg