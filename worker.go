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

	go w.read()
	go w.write()

	w.Swarm.Logchan.Info <- "Worker [" + id + "] created."

	return w
}

// Read channel for any incoming messages from the worker machine
func (w *Worker) read() {
	var (
		id   = w.ID
		data string
		err  error
	)

	w.Swarm.Logchan.Info <- "Worker [" + id + "] started listening to read channel."

	for {
		w.Swarm.Logchan.Info <- "Worker [" + id + "] listening to read..."

		if w.Conn == nil {
			w.Swarm.Logchan.Warning <- "Worker [" + id + "] connection was destroyed.[1]"
			break
		}
		w.Swarm.Logchan.Info <- "r"
		data, err = w.Reader.ReadString('\n')

		if w.Conn == nil {
			w.Swarm.Logchan.Warning <- "Worker [" + id + "] connection was destroyed.[2]"
			break
		}

		if err == nil {
			w.Swarm.Logchan.Message <- string(log.NewMessage(id, data, true).JSON())
			w.Swarm.ReadData <- Message{Content: data, Worker: w, Read: true} // sends message to swarm
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

	w.Swarm.Logchan.Info <- "Worker [" + id + "] read channel closed."
}

// Write channel for sending messages to the worker machine
func (w *Worker) write() {
	var id = w.ID

	w.Swarm.Logchan.Info <- "Worker [" + id + "] started listening to write channel."

loop:
	for {
		w.Swarm.Logchan.Info <- "Worker [" + id + "] listening to write..."
		select {
		case <-w.killedWrite:
			w.Swarm.Logchan.Info <- "wx"
			break loop

		case data := <-w.WriteData:
			w.Swarm.Logchan.Info <- "w"
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
