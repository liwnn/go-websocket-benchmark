package connections

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go-websocket-benchmark/config"
	"go-websocket-benchmark/logging"
	"go-websocket-benchmark/mwsbench/report"

	"github.com/lesismal/nbio"
	nblog "github.com/lesismal/nbio/logging"
	"github.com/lesismal/perf"
)

type Connections struct {
	Framework      string
	Ip             string
	Concurrency    int
	NumConnections int
	DialTimeout    time.Duration
	RetryInterval  time.Duration
	RetryTimes     int
	Percents       []int

	// Caculations
	Success uint32
	Failed  uint32

	// All connected connections
	connsMap map[*nbio.Conn]struct{}

	Calculator *perf.Calculator

	Engine *nbio.Engine

	mux          sync.Mutex
	serverIdx    uint32
	serverAddrs  []string
	chConnecting chan struct{}
}

func New(framework, ip string, numConns int) *Connections {
	return &Connections{
		Framework:      framework,
		Ip:             ip,
		NumConnections: numConns,
		connsMap:       map[*nbio.Conn]struct{}{},
	}
}

func (cs *Connections) Run() {
	cs.init()
	defer cs.clean()

	logging.Printf("Dial Connections: [%v]", cs.NumConnections)
	logging.Printf("Dial Concurrency: [%v]", cs.Concurrency)
	done := make(chan struct{})
	logCone := make(chan struct{})

	go func() {
		defer func() {
			logging.Printf("Connections done: %v Success, %v Failed", cs.Success, cs.Failed)
			close(logCone)
		}()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for i := 1; true; i++ {
			select {
			case <-done:
				return
			case <-ticker.C:
				logging.Printf("%03d secons passed, %v Connected ...", i, atomic.LoadUint32(&cs.Success))
			}
		}
	}()

	logging.Printf("Connections start ...")
	cs.Calculator.Benchmark(cs.Concurrency, cs.NumConnections, cs.doOnce, cs.Percents)

	close(done)
	<-logCone
}

func (cs *Connections) Conns() map[*nbio.Conn]struct{} {
	return cs.connsMap
}

func (cs *Connections) Stop() {
	for c := range cs.connsMap {
		c.Close()
	}
	cs.Engine.Shutdown(context.Background())
}

func (cs *Connections) Report() report.Report {
	return &report.ConnectionsReport{
		Framework:   cs.Framework,
		TPS:         cs.Calculator.TPS(),
		Min:         cs.Calculator.Min,
		Avg:         cs.Calculator.Avg,
		Max:         cs.Calculator.Max,
		TP50:        cs.Calculator.TPN(50),
		TP75:        cs.Calculator.TPN(75),
		TP90:        cs.Calculator.TPN(90),
		TP95:        cs.Calculator.TPN(95),
		TP99:        cs.Calculator.TPN(99),
		Used:        int64(cs.Calculator.Used),
		Total:       cs.NumConnections,
		Success:     cs.Success,
		Failed:      cs.Failed,
		Concurrency: cs.Concurrency,
	}
}

func (cs *Connections) init() {
	if cs.NumConnections <= 0 {
		cs.NumConnections = 1000
	}
	if cs.Concurrency <= 0 {
		cs.Concurrency = runtime.NumCPU() * 1000
	}
	if cs.Concurrency > cs.NumConnections {
		cs.Concurrency = cs.NumConnections
	}
	if cs.DialTimeout <= 0 {
		cs.DialTimeout = time.Second * 1
	}
	if cs.RetryInterval <= 0 {
		cs.RetryInterval = time.Second / 10
	}
	if cs.RetryTimes <= 0 {
		cs.RetryTimes = 3
	}
	if len(cs.Percents) == 0 {
		cs.Percents = []int{50, 75, 90, 95, 99}
	}

	cs.Calculator = perf.NewCalculator(fmt.Sprintf("%v-Connect", cs.Framework))

	ports, err := config.GetFrameworkBenchmarkPorts(cs.Framework)
	var addrs []string
	for _, port := range ports {
		addrs = append(addrs, fmt.Sprintf("%v:%v", cs.Ip, port))
	}
	if err != nil {
		logging.Fatalf("GetFrameworkBenchmarkAddrs(%v) failed: %v", cs.Framework, err)
	}
	cs.serverAddrs = addrs

	cs.chConnecting = make(chan struct{}, cs.NumConnections)
	for i := 0; i < cs.NumConnections; i++ {
		cs.chConnecting <- struct{}{}
	}

	cs.startEngine()
}

func (cs *Connections) clean() {
	cs.serverAddrs = nil
	cs.chConnecting = nil
}

func (cs *Connections) startEngine() {
	if cs.Engine != nil {
		return
	}

	nblog.Output = logging.Output
	nblog.SetLevel(nblog.LevelError)

	engine := nbio.NewEngine(nbio.Config{Name: "Benchmark-TCP-Client"})
	err := engine.Start()
	if err != nil {
		logging.Fatalf("nbhttp.Engine.Start failed: %v\n", err)
	}
	engine.OnData(func(c *nbio.Conn, data []byte) {})
	cs.Engine = engine

	time.Sleep(time.Second)
}

func (cs *Connections) doOnce() error {
begin:
	for {
		select {
		case <-cs.chConnecting:
		default:
			return nil
		}

		for i := 0; i < cs.RetryTimes; i++ {
			addr := cs.serverAddrs[atomic.AddUint32(&cs.serverIdx, 1)%uint32(len(cs.serverAddrs))]
			conn, err := nbio.DialTimeout("tcp", addr, cs.DialTimeout)
			if err == nil {
				cs.Engine.AddConn(conn)
				conn.SetReadDeadline(time.Time{})
				atomic.AddUint32(&cs.Success, 1)
				cs.mux.Lock()
				cs.connsMap[conn] = struct{}{}
				cs.mux.Unlock()
				goto begin
			}
			time.Sleep(cs.RetryInterval)
		}
		atomic.AddUint32(&cs.Failed, 1)
	}
}
