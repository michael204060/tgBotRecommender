package events

type Type int

type Event struct {
	Type Type
	Text string
	Meta interface{}
}

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Processor interface {
	Process(e Event) error
}

const (
	Unknown Type = iota
	Message
	Callback
)
