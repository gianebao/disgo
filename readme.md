[![Go Report Card](https://goreportcard.com/badge/github.com/matchmove/rest)](https://goreportcard.com/report/github.com/gianebao/disgo)
[![GoDoc](https://godoc.org/github.com/matchmove/rest?status.svg)](https://godoc.org/github.com/gianebao/disgo)

  import "gopkg.in/gianebao/disgo"

# disgo
--
    import "github.com/gianebao/disgo"

## Usage

#### func  MakeWorkerID

```go
func MakeWorkerID(a net.Addr) string
```
MakeWorkerID creates an ID from net.Addr

#### type Message

```go
type Message struct {
	Content string
	Worker  *Worker
	Read    bool
}
```

Message represents the message received by the a worker

#### type Swarm

```go
type Swarm struct {
	Workers       map[string]*Worker
	NewConnection chan net.Conn

	Logchan *log.Channel
	Mux     sync.Mutex
}
```

Swarm represents the group of the workers

#### func  NewSwarm

```go
func NewSwarm(l *log.Channel) *Swarm
```
NewSwarm creates a new swarm instance

#### func (*Swarm) HandleNewConnections

```go
func (s *Swarm) HandleNewConnections(fn func(c net.Conn) bool) *Swarm
```
HandleNewConnections listens to instructions for NewConnection channels

#### func (*Swarm) Kill

```go
func (s *Swarm) Kill(w *Worker)
```
Kill the worker in the swarm

#### func (*Swarm) NewWorker

```go
func (s *Swarm) NewWorker(c net.Conn) *Worker
```
NewWorker creates a new worker in the swarm

#### func (*Swarm) Reader

```go
func (s *Swarm) Reader(fn func(m Message) string) *Swarm
```
Reader listens to instructions for ReadData channels.

#### func (*Swarm) Writer

```go
func (s *Swarm) Writer(fn func(m Message) bool) *Swarm
```
Writer listens to instructions for WriteData channels.

#### type Worker

```go
type Worker struct {
	ID        string
	Addr      net.Addr
	Swarm     *Swarm
	Conn      net.Conn
	ReadData  chan string
	WriteData chan string
	Reader    *bufio.Reader
	Writer    *bufio.Writer

	Mux sync.Mutex
}
```

Worker represents the machine connected to a swarm

#### func  NewWorker

```go
func NewWorker(id string, s *Swarm, c net.Conn) *Worker
```
NewWorker creates a new worker instance

#### func (*Worker) Die

```go
func (w *Worker) Die()
```
Die destroys the worker

## Example

## Logging Channel


Documentation build by [robertkrimen/godocdown](https://github.com/robertkrimen/godocdown)
