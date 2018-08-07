package logopher

import "time"

const(
	EMERGENCY = "EMERGENCY"
	CRITICAL = "CRITICAL"
	ERROR = "ERROR"
	WARNING = "WARNING"
	NOTICE = "NOTICE"
	INFO = "INFO"
	DEBUG = "DEBUG"
)

type MessageContext map[string]interface{}

type Message struct {
	Message   string
	Timestamp *time.Time
	Level     string
	Context   MessageContext
}

func (m *Message) Clone() *Message {
	newMessage := Message{
		Message: m.Message,
		Level: m.Level,
	}
	timestampVal := *m.Timestamp
	newMessage.Timestamp = &timestampVal

	contextVal := make(MessageContext)
	for key, val := range m.Context {
		contextVal[key] = val
	}
	newMessage.Context = contextVal

	return &newMessage
}

func createMessageFromString(level, message string, context *MessageContext) *Message{
	now := time.Now()

	var contextVal MessageContext
	if nil == context {
		contextVal = make(MessageContext)
	} else {
		contextVal = *context
	}

	return &Message{
		Message:   message,
		Level:     level,
		Timestamp: &now,
		Context:   contextVal,
	}
}