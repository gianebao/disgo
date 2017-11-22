package disgo

// Message represents the message received by the a worker
type Message struct {
	Content string
	Worker  *Worker
	Read    bool
}
