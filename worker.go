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
	ID        string
	Addr      net.Addr
	Swarm     *Swarm
	Conn      net.Conn
	ReadData  chan string
	WriteData chan string
	Reader    *bufio.Reader
	Writer    *bufio.Writer
	killed    chan bool
	Mux       sync.Mutex
}

// NewWorker creates a new worker instance
func NewWorker(id string, s *Swarm, c net.Conn) *Worker {
	writer := bufio.NewWriter(c)
	reader := bufio.NewReader(c)

	w := &Worker{
		ID:        id,
		Swarm:     s,
		Conn:      c,
		ReadData:  make(chan string),
		WriteData: make(chan string),
		Reader:    reader,
		Writer:    writer,
		killed:    make(chan bool),
	}

	go w.read()
	go w.write()

	s.Logchan.Info <- "Worker [" + id + "] created. Start listening to io channel."

	return w
}

// Read channel for any incoming messages from the worker machine
func (w *Worker) read() {
	var (
		data string
		err  error
	)

	for {
		select {
		case <-w.killed:
			return

		default:
			data, err = w.Reader.ReadString('\n')

			if nil == w.Conn {
				return
			}

			if err == nil {
				w.Swarm.Logchan.Message <- string(log.NewMessage(w.ID, data, true).JSON())
				w.Swarm.ReadData <- Message{Content: data, Worker: w, Read: true} // sends message to swarm
				continue
			}

			// EOF indicates disconnection
			if err == io.EOF {
				w.Swarm.Logchan.Warning <- "Worker [" + w.ID + "] was disconnected. Starting \"kill\" process..."
				w.Die()
				return
			}

			w.Swarm.Logchan.Error <- fmt.Sprintf("Worker [%s] cannot read with error [%v]", w.ID, err)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Write channel for sending messages to the worker machine
func (w *Worker) write() {
	for {
		select {
		case <-w.killed:
			return
		case data := <-w.WriteData:
			if nil == w.Conn {
				return
			}

			w.Swarm.Logchan.Message <- string(log.NewMessage(w.ID, data, false).JSON())
			w.Writer.WriteString(data)
			w.Writer.Flush()
		}
	}
}

// Die destroys the worker
func (w *Worker) Die() {
	w.Swarm.Kill(w)
}
