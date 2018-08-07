package logopher

import (
	"fmt"
	"github.com/bassbeaver/logopher/dateformat"
	"encoding/json"
)

type FormatterInterface interface {
	Format(message *Message) string
}

// -----------------------------------

type SimpleFormatter struct {}

func (f *SimpleFormatter) Format(message *Message) string {
	return fmt.Sprintf(
		"%s [%s] %s",
		message.Timestamp.Format(dateformat.DateTimeMicrosecFormat),
		message.Level,
		message.Message,
	)
}

// -----------------------------------

type JsonFormatter struct {}

func (f *JsonFormatter) Format(message *Message) string {
	jsonString, jsonError := json.Marshal(map[string]interface{}{
		"timestamp": message.Timestamp.Format(dateformat.DateTimeMicrosecFormat),
		"level": message.Level,
		"message": message.Message,
		"context": message.Context,
	})
	if jsonError != nil {
		return ""
	}

	return string(jsonString)
}