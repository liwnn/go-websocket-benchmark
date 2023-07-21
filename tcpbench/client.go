package main

import (
	"flag"
	"runtime/debug"
	"time"

	"go-websocket-benchmark/config"
	"go-websocket-benchmark/logging"
	"go-websocket-benchmark/mwsbench/report"
	"go-websocket-benchmark/tcpbench/connections"
)

var (
	// Client Proc
	memLimit = flag.Int64("m", 1024*1024*1024*4, `memory limit`)

	// Server Side
	framework = flag.String("f", config.NbioTcp, `framework, e.g. "gorilla"`)
	ip        = flag.String("ip", "127.0.0.1", `ip, e.g. "127.0.0.1"`)

	// Connection
	numConnections    = flag.Int("c", 10000, "client: num of connections")
	dialConcurrency   = flag.Int("dc", 2000, "client: dial concurrency: how many goroutines used to do dialing")
	dialTimeout       = flag.Duration("dt", 5*time.Second, "client: dial timeout")
	dialRetries       = flag.Int("dr", 5, "client: dial retry times")
	dialRetryInterval = flag.Duration("dri", 100*time.Millisecond, "client; dial retry interval")

	// BenchEcho && BenchRate
	payload = flag.Int("b", 1024, `benchmark: payload size of benchecho and benchrate`)

	// BenchEcho
	echoConcurrency = flag.Int("ec", 50000, "benchecho: concurrency: how many goroutines used to do the echo test")
	echoTimes       = flag.Int("en", 2000000, `benchecho: benchmark times`)
	echoTPSLimit    = flag.Int("el", 0, `benchecho: TPS limitation per second`)

	// BenchRate
	rateEnabled     = flag.Bool("rate", false, `benchrate: whether run benchrate`)
	rateConcurrency = flag.Int("rc", 50000, "benchrate: concurrency: how many goroutines used to do the echo test")
	rateDuration    = flag.Int("rd", 10, `benchrate: how long to spend to do the test`)
	rateSendRate    = flag.Int("rr", 100, "benchrate: how many request message can be sent to 1 conn every second")
	rateSendLimit   = flag.Int("rl", 0, `benchrate: message sending limitation per second`)

	// for report generation
	genReport = flag.Bool("r", false, `make report`)
	preffix   = flag.String("preffix", "", `report file preffix, e.g. "1m_connections_"`)
	suffix    = flag.String("suffix", "", `report file suffix, e.g. "_20060102150405"`)
)

func main() {
	flag.Parse()

	if *genReport {
		generateReports()
		return
	}

	debug.SetMemoryLimit(*memLimit)

	logging.Print(logging.LongLine)
	defer logging.Print(logging.LongLine)

	logging.Printf("Benchmark [%v]: %v connections, %v payload, %v times", *framework, *numConnections, *payload, *echoTimes)
	logging.Print(logging.ShortLine)

	cs := connections.New(*framework, *ip, *numConnections)
	cs.Concurrency = *dialConcurrency
	cs.DialTimeout = *dialTimeout
	cs.RetryTimes = *dialRetries
	cs.RetryInterval = *dialRetryInterval
	cs.Run()
	defer cs.Stop()
	csReport := cs.Report()
	report.ToFile(csReport, *preffix, *suffix)
	logging.Print(logging.ShortLine)
	logging.Print(csReport.String())
	logging.Print("\n")
	logging.Print(logging.ShortLine)
}

func generateReports() {
	data := report.GenerateConnectionsReports(*preffix, *suffix, nil)
	filename := report.Filename("Connections", *preffix, *suffix+".md")
	report.WriteFile(filename, data)
	logging.Print(logging.LongLine)
	logging.Printf("[%vConnections%v] Report\n", *preffix, *suffix)
	logging.Print(data)
}
