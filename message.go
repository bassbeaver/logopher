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
	message string
	timestamp *time.Time
	level string
	context *MessageContext
}

func createMessageFromString(level, message string, context *MessageContext) *Message{
	now := time.Now()

	return &Message{
		message: message,
		level: level,
		timestamp: &now,
		context: context,
	}
}