package logopher

import (
	"fmt"
	"logopher/dateformat"
	"encoding/json"
)

type FormatterInterface interface {
	format(message *Message) string
}

// -----------------------------------

type SimpleFormatter struct {}

func (f *SimpleFormatter) format(message *Message) string {
	return fmt.Sprintf(
		"%s [%s] %s",
		message.timestamp.Format(dateformat.DateTimeFormat),
		message.level,
		message.message,
	)
}

// -----------------------------------

type JsonFormatter struct {}

func (f *JsonFormatter) format(message *Message) string {
	jsonString, jsonError := json.Marshal(map[string]interface{}{
		"timestamp": message.timestamp.Format(dateformat.DateTimeFormat),
		"level": message.level,
		"message": message.message,
		"context": message.context,
	})
	if jsonError != nil {
		return ""
	}

	return string(jsonString)
}