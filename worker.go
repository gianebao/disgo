package disgo

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gianebao/disgo/log"
)

// Worker represents the machine connected to a swarm
type Worker struct {
	ID          string
	Addr        net.Addr
	Swarm       *Swarm
	Conn        net.Conn
	ReadData    chan string
	WriteData   chan string
	Reader      *bufio.Reader
	Writer      *bufio.Writer
	killedRead  chan struct{}
	killedWrite chan struct{}
	killedConn  chan struct{}
	Mux         sync.Mutex
}

// NewWorker creates a new worker instance
func NewWorker(id string, s *Swarm, c net.Conn) *Worker {
	writer := bufio.NewWriter(c)
	reader := bufio.NewReader(c)

	w := &Worker{
		ID:          id,
		Swarm:       s,
		Conn:        c,
		ReadData:    make(chan string),
		WriteData:   make(chan string),
		Reader:      reader,
		Writer:      writer,
		killedRead:  make(chan struct{}),
		killedWrite: make(chan struct{}),
		killedConn:  make(chan struct{}),
	}

	go w.buff()
	go w.read()
	go w.write()

	w.Swarm.Logchan.Info <- "Worker [" + id + "] created."

	return w
}

// buff reads input from connection and forwards to read channel
func (w *Worker) buff() {
	var (
		id   = w.ID
		data string
		err  error
	)

	w.Swarm.Logchan.Info <- "Worker [" + id + "] started listening to buff channel."

	for {
		w.Swarm.Logchan.Info <- "Worker [" + id + "] listening to buff..."
		data, err = w.Reader.ReadString('\n')

		if w.Conn == nil {
			w.Swarm.Logchan.Warning <- "Worker [" + id + "] connection was destroyed.[1]"
			break
		}

		if err == nil {
			w.ReadData <- data
			continue
		}

		// EOF indicates disconnection
		if err == io.EOF {
			w.Swarm.Logchan.Warning <- "Worker [" + id + "] was disconnected."
			w.Die()
			break
		}

		w.Swarm.Logchan.Error <- fmt.Sprintf("Worker [%s] cannot read with error [%v].", id, err)
		time.Sleep(100 * time.Millisecond)
	}

	w.Swarm.Logchan.Info <- "Worker [" + id + "] buff closed."
}

// read channel for any incoming messages from the worker machine
func (w *Worker) read() {
	var id = w.ID

	w.Swarm.Logchan.Info <- "Worker [" + id + "] started listening to read channel."

loop:
	for {
		w.Swarm.Logchan.Info <- "Worker [" + id + "] listening to read..."
		select {
		case <-w.killedRead:
			break loop

		case data := <-w.ReadData:
			w.Swarm.Logchan.Message <- string(log.NewMessage(id, data, true).JSON())
			data = w.Swarm.reader(Message{Content: data, Worker: w, Read: true})
			if w.Conn == nil {
				break loop
			}
			if data != "" {
				w.WriteData <- data
			}
		}
	}

	w.Swarm.Logchan.Info <- "Worker [" + id + "] read channel closed."
}

// write channel for sending messages to the worker machine
func (w *Worker) write() {
	var id = w.ID

	w.Swarm.Logchan.Info <- "Worker [" + id + "] started listening to write channel."

loop:
	for {
		w.Swarm.Logchan.Info <- "Worker [" + id + "] listening to write..."
		select {
		case <-w.killedWrite:
			break loop

		case data := <-w.WriteData:
			w.Swarm.Logchan.Message <- string(log.NewMessage(id, data, false).JSON())
			w.Writer.WriteString(data)
			w.Writer.Flush()
		}
	}

	w.Swarm.Logchan.Info <- "Worker [" + id + "] write channel closed."
}

// Die destroys the worker
func (w *Worker) Die() {
	w.Swarm.Kill(w)
}
