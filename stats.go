package main

import (
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	influx "github.com/influxdata/influxdb/client/v2"
)

type AccumulatedStats struct {
	since          time.Time
	submittedTasks int
	completedTasks int
	failedTasks    int
}

type Stats struct {
	imp          chan interface{}
	mem          runtime.MemStats
	client       influx.Client
	influxConfig influx.HTTPConfig
	accumulated  AccumulatedStats
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

func (s *Stats) pulse() {
	s.imp <- nil
}

func (s *Stats) reportRuntimeStats() {
	runtime.ReadMemStats(&s.mem)
	if s.mem.LastGC > 0 && false {
		log.Println()
		log.Println(s.mem.Alloc)
		log.Println(s.mem.HeapSys)
		log.Println(s.mem.LastGC)
		log.Println(s.mem.NextGC)
	}

	tags := map[string]string{"mem": "memory"}
	fields := map[string]interface{}{
		"total_alloc": s.mem.TotalAlloc,
		"alloc":       s.mem.Alloc,
		"heap_sys":    s.mem.HeapSys,
		"next_gc":     s.mem.NextGC,
	}
	pt, err := influx.NewPoint("memory", tags, fields, time.Now())
	s.bp.AddPoint(pt)
	if err != nil {
		log.Error(err)
	}

	pt, err = influx.NewPoint(
		"goroutines",
		map[string]string{"goroutines": "number-of-goroutines"},
		map[string]interface{}{"goroutines": runtime.NumGoroutine()},
		time.Now(),
	)
	s.bp.AddPoint(pt)
	if err != nil {
		log.Error(err)
	}

	pt, err = influx.NewPoint(
		"next-task",
		map[string]string{"next-task": "next-task"},
		map[string]interface{}{"next_task": currTaskID},
		time.Now(),
	)
	s.bp.AddPoint(pt)
	if err != nil {
		log.Error(err)
	}

}

func (s *Stats) reportAccumulated() {
	s.accumulated = AccumulatedStats{since: time.Now()}
}

func (s *Stats) sendPeriodicalReports() {
	idleTicker := time.NewTicker(time.Minute)
	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			active := false
			for w := range workers() {
				active = active || w.active
			}
			if !active {
				continue
			}
			log.Debug("ticktock, active!")
		case <-idleTicker.C:
			log.Debug("ticktock, idle!")
		case <-s.imp:
			drainCh(s.imp)
		}
		s.reportRuntimeStats()
		s.reportAccumulated()
		err := s.client.Write(s.bp)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

var stats Stats

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
		stats.accumulated = AccumulatedStats{since: time.Now()}
		stats.reportRuntimeStats()
		stats.reportAccumulated()
		err = stats.client.Write(stats.bp)
		if err != nil {
			log.Error(err)
			return
		}
		go stats.sendPeriodicalReports()
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
