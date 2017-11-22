package log

// Channel represents the log messages created by the hive
type Channel struct {
	Fatal   chan string
	Info    chan string
	Warning chan string
	Error   chan string
	Message chan string
}

// NewChannel creates a new Channel instance
func NewChannel() *Channel {
	return &Channel{
		Fatal:   make(chan string),
		Info:    make(chan string),
		Warning: make(chan string),
		Error:   make(chan string),
		Message: make(chan string),
	}
}
