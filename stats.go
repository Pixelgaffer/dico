package main

import (
	"runtime"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	influx "github.com/influxdata/influxdb/client/v2"
)

type AccumulatedStats struct {
	sync.Mutex
	since          time.Time
	submittedTasks int
	completedTasks int
	failedTasks    int
}

func (s *AccumulatedStats) Reset() {
	s.Lock()
	s.since = time.Now()
	s.submittedTasks = 0
	s.completedTasks = 0
	s.failedTasks = 0
	s.Unlock()
}

func (s *AccumulatedStats) With(f func(s *AccumulatedStats)) {
	s.Lock()
	f(s)
	s.Unlock()
}

type Stats struct {
	imp          chan interface{}
	mem          runtime.MemStats
	client       influx.Client
	influxConfig influx.HTTPConfig
	accumulated  *AccumulatedStats
	bp           influx.BatchPoints
}

func (s *Stats) initClient() (err error) {
	log.WithField("Addr", s.influxConfig.Addr).Debug("connecting to influxdb")
	s.client, err = influx.NewHTTPClient(s.influxConfig)
	if err != nil {
		return err
	}
	defer s.client.Close()
	ping, _, err := s.client.Ping(time.Second)
	if err != nil {
		return err
	}
	log.WithField("ping", ping).Debug("influx ping")

	s.bp, err = influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  "dicod",
		Precision: "s",
	})
	return err
}

func (s *Stats) Pulse() {
	s.imp <- nil
}

func (s *Stats) reportRuntimeStats() {
	runtime.ReadMemStats(&s.mem)

	s.newPoint(
		"memory",
		"total_alloc", s.mem.TotalAlloc,
		"alloc", s.mem.Alloc,
		"heap_sys", s.mem.HeapSys,
		"next_gc", s.mem.NextGC,
	)

	s.newPoint("goroutines", runtime.NumGoroutine())
	s.newPoint("next_task", currTaskID)
}

func (s *Stats) reportAccumulated() {
	s.accumulated.Lock()
	//sinceSec := float64(s.accumulated.since.Unix())
	s.newPoint(
		"tasks",
		"submitted", s.accumulated.submittedTasks,
		"completed", s.accumulated.completedTasks,
		"failed", s.accumulated.failedTasks,
	)
	s.accumulated.Unlock()
	s.accumulated.Reset()
}

func (s *Stats) newPoint(name string, values ...interface{}) {
	var fields map[string]interface{}
	if len(values) == 1 {
		fields = map[string]interface{}{name: values[0]}
	} else if len(values)%2 == 0 {
		fields = make(map[string]interface{})
		for i := 0; i < len(values)/2; i++ {
			fields[values[i*2].(string)] = values[i*2+1]
		}
	} else {
		panic("invalid number of args")
	}
	pt, err := influx.NewPoint(
		strings.Replace(name, "_", "-", -1),
		map[string]string{},
		fields,
		time.Now(),
	)
	s.bp.AddPoint(pt)
	if err != nil {
		log.Error(err)
	}
}

func (s *Stats) sendPeriodicalReports() {
	idleTicker := time.NewTicker(time.Minute)
	ticker := time.NewTicker(time.Second)
	for {
		isImpulse := false
		select {
		case <-ticker.C:
			anyActive := false
			for w := range workers() {
				anyActive = anyActive || w.active
			}
			if !anyActive {
				continue
			}
			log.Debug("ticktock, active!")
		case <-idleTicker.C:
			log.Debug("ticktock, idle!")
		case <-s.imp:
			isImpulse = true
			drainCh(s.imp)
		}
		var activeWorkers int
		var numWorkers int
		for w := range workers() {
			numWorkers++
			if w.active {
				activeWorkers++
			}
		}
		s.reportRuntimeStats()
		if !isImpulse {
			s.reportAccumulated()
		}
		s.newPoint("workers", "workers", numWorkers, "active", activeWorkers)
		err := s.client.Write(s.bp)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

var stats Stats
var accumulatedStats AccumulatedStats

func init() {
	go func() {
		stats = Stats{
			imp: make(chan interface{}, 50),
			influxConfig: influx.HTTPConfig{ // TODO: read from config file
				Addr:     "http://influx.thuermchen.com",
				Username: "dicod",
				Password: "dicod",
			},
		}
		err := stats.initClient()
		if err != nil {
			log.Error(err)
			return
		}
		accumulatedStats.Reset()
		stats.accumulated = &accumulatedStats
		err = stats.client.Write(stats.bp)
		if err != nil {
			log.Error(err)
			return
		}
		go stats.sendPeriodicalReports()
		stats.Pulse()
	}()
}

func drainCh(c chan interface{}) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}
