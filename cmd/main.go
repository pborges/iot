package main

import (
	"errors"
	"github.com/pborges/iot/process"
	"github.com/pborges/iot/pubsub"
	"sync"
)

type Server struct {
	broker    pubsub.Broker
	processes map[string]process.Process
	lock      sync.Mutex
}

func (s Server) CreateProcess(name string) (process.Process, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.processes == nil {
		s.processes = make(map[string]process.Process)
	}
	if _, ok := s.processes[name]; !ok {
		return process.Process{}, errors.New("process with this name already exists")
	}
	proc := process.Process{
		Broker: s.broker,
	}
	s.processes[name] = proc
	return proc, nil
}

func main() {
}
