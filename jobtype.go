package main

import (
	"sync"

	protos "github.com/Pixelgaffer/dico-proto"
)

var jobTypes struct {
	all []*JobType
	sync.Mutex
}

type JobType struct {
	archive []byte
	hash    string
	name    string
	proto   *protos.SubmitCode
}

func sendJobTypes(w *Worker) {
	jobTypes.Lock()
	for _, jt := range jobTypes.all {
		w.connection.send <- jt.proto
	}
	jobTypes.Unlock()
}

func addJobType(p *protos.SubmitCode) {
	jt := &JobType{
		name:    p.GetJobType(),
		hash:    p.GetHash(),
		archive: p.GetArchive(),
		proto:   p,
	}
	jobTypes.Lock()
	jobTypes.all = append(jobTypes.all, jt)
	jobTypes.Unlock()
	for worker := range workers() {
		worker.connection.send <- p
	}
}
