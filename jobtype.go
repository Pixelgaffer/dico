package main

import (
	"sync"

	protos "github.com/Pixelgaffer/dico-proto"
)

var jobTypes []*JobType
var jobTypesLock sync.Mutex // TODO: struct

type JobType struct {
	archive []byte
	hash    string
	name    string
	proto   *protos.SubmitCode
}

func sendJobTypes(w *Worker) {
	jobTypesLock.Lock()
	for _, jt := range jobTypes {
		w.connection.send <- jt.proto
	}
	jobTypesLock.Unlock()
}

func addJobType(p *protos.SubmitCode) {
	jt := &JobType{
		name:    p.GetJobType(),
		hash:    p.GetHash(),
		archive: p.GetArchive(),
		proto:   p,
	}
	jobTypesLock.Lock()
	jobTypes = append(jobTypes, jt)
	jobTypesLock.Unlock()
	for worker := range workers() {
		worker.connection.send <- p
	}
}
