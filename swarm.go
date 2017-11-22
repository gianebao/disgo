package disgo

import (
	"net"
	"sync"

	"github.com/gianebao/disgo/log"
)

// Swarm represents the group of the workers
type Swarm struct {
	Workers       map[string]*Worker
	NewConnection chan net.Conn
	ReadData      chan Message
	WriteData     chan Message
	Logchan       *log.Channel
	Mux           sync.Mutex
}

// MakeWorkerID creates an ID from net.Addr
func MakeWorkerID(a net.Addr) string {
	return a.Network() + "-" + a.String()
}

// NewSwarm creates a new swarm instance
func NewSwarm(l *log.Channel) *Swarm {
	s := &Swarm{
		Workers:       make(map[string]*Worker),
		NewConnection: make(chan net.Conn),
		ReadData:      make(chan Message),
		WriteData:     make(chan Message),
		Logchan:       l,
	}

	s.Logchan.Info <- "Swarm created."
	return s
}

// NewWorker creates a new worker in the swarm
func (s *Swarm) NewWorker(c net.Conn) *Worker {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	id := MakeWorkerID(c.RemoteAddr())

	if _, keyExists := s.Workers[id]; keyExists {
		s.Logchan.Error <- "Connection [" + id + "] already exists in the swarm. Ignoring connection!"
		return nil
	}

	w := NewWorker(id, s, c)
	s.Workers[id] = w
	s.Logchan.Info <- "Worker [" + id + "] was successfully added to the swarm."

	return w
}

// HandleRead listens to instructions for ReadData channels.
func (s *Swarm) HandleRead(fn func(m Message) string) *Swarm {
	go func(s *Swarm) {
		for {
			select {
			// catch all incoming messages. Executed after the the worker.
			case m := <-s.ReadData:
				if nil != fn {
					m.Worker.WriteData <- fn(m)
				}
			}
		}
	}(s)

	return s
}

// HandleWrite listens to instructions for WriteData channels.
func (s *Swarm) HandleWrite(fn func(m Message) bool) *Swarm {
	go func(s *Swarm) {
		for {
			select {
			// catch all outgoing messages.
			case m := <-s.WriteData:
				if nil != fn && !fn(m) {
					return
				}
			}
		}
	}(s)

	return s
}

// HandleNewConnections listens to instructions for NewConnection channels
func (s *Swarm) HandleNewConnections(fn func(c net.Conn) bool) *Swarm {
	go func(s *Swarm) {
		for {
			select {
			case conn := <-s.NewConnection:
				if nil == fn || fn(conn) {
					s.NewWorker(conn)
				}
			}
		}
	}(s)

	return s
}

// Kill the worker in the swarm
func (s *Swarm) Kill(w *Worker) {
	s.Mux.Lock()
	w.Mux.Lock()

	defer func() {
		s.Mux.Unlock()
		w.Mux.Unlock()
	}()

	delete(s.Workers, w.ID)
	w.killed <- true
	w.Conn.Close()
	w.Conn = nil

	s.Logchan.Warning <- "Successfully killed worker [" + w.ID + "] and removed from the swarm."
}
