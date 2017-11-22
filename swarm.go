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
	reader        func(m Message) string
	writer        func(m Message) bool
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
		reader:        func(m Message) string { return "" },
		writer:        func(m Message) bool { return true },
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

// Reader listens to instructions for ReadData channels.
func (s *Swarm) Reader(fn func(m Message) string) *Swarm {
	s.reader = fn

	return s
}

// Writer listens to instructions for WriteData channels.
func (s *Swarm) Writer(fn func(m Message) bool) *Swarm {
	s.writer = fn

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
	id := w.ID

	s.Logchan.Info <- "Starting \"kill\" for worker [" + id + "]..."

	s.Mux.Lock()
	w.Mux.Lock()

	defer func() {
		s.Mux.Unlock()
		w.Mux.Unlock()
		w = nil
	}()

	if _, keyExists := s.Workers[id]; keyExists {
		delete(s.Workers, id)
	}

	if nil != w.Conn {
		w.Conn.Close()
		w.Conn = nil
	}

	close(w.killedWrite)
	s.Logchan.Warning <- "Successfully killed worker [" + id + "] and removed from the swarm."
}
