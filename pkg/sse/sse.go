package sse

type ServerSentEvent struct {
	Data     *string
	Event    *string
	ID       *string
	Retry    *int
	Comments *string
}

func NewServerSentEvent() *ServerSentEvent {
	return &ServerSentEvent{}
}

// On the client, an event listener can be setup to listen to that particular event.
func (s *ServerSentEvent) WithEvent(event string) *ServerSentEvent {
	s.Event = &event
	return s
}

// Setting an ID lets the browser keep track of the last event fired so that if,
// the connection to the server is dropped, a special HTTP header (Last-Event-ID) is
// set with the new request. This lets the browser determine which event is appropriate to fire.
// The message event contains a e.lastEventId property.
func (s *ServerSentEvent) WithID(id string) *ServerSentEvent {
	s.ID = &id
	return s
}

// WithRetry sets the reconnection time in milliseconds.
// by default browsers will try to reconnect to the server after 3 seconds
// after each session is closed.
func (s *ServerSentEvent) WithRetry(retry int) *ServerSentEvent {
	s.Retry = &retry
	return s
}

func (s *ServerSentEvent) WithComments(comments string) *ServerSentEvent {
	s.Comments = &comments
	return s
}

func (s *ServerSentEvent) WithData(data string) *ServerSentEvent {
	s.Data = &data
	return s
}
