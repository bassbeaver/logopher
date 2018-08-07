package logopher

import (
	"testing"
	"time"
	"reflect"
)

func TestMessageCloning(t *testing.T) {
	now := time.Now()
	message1 := &Message{
		Message:   "message1",
		Level:     INFO,
		Timestamp: &now,
		Context:   MessageContext{"field1": "value1"},
	}

	message2 := message1.Clone()

	message1.Message = "message1_updated"
	message1.Level = WARNING
	*message1.Timestamp = message1.Timestamp.AddDate(0, 0, 1)
	message1.Context["field1"] = "value1_updated"

	if
		reflect.DeepEqual(message1.Message, message2.Message) ||
		reflect.DeepEqual(message1.Level, message2.Level) ||
		reflect.DeepEqual(message1.Timestamp, message2.Timestamp) ||
		reflect.DeepEqual(message1.Context, message2.Context) {

		t.Errorf(
			"Message cloning failed. Changes to message1 propagated to message2 \n"+
			"message1: {Message: %s, \t Timestamp: %v, \t\t\t\t Level: %s, \t Context: %v} \n"+
			"message2: {Message: %s, \t\t\t Timestamp: %v, Level: %s, \t\t Context: %v} \n",
			message1.Message, message1.Timestamp, message1.Level, message1.Context,
			message2.Message, message2.Timestamp, message2.Level, message2.Context,
		)
	}
}