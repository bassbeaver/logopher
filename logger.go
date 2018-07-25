package logopher

type Logger struct {
	handlers []HandlerInterface
}

func (l *Logger) Log(level, message string, context *MessageContext) {
	messageObj := createMessageFromString(level, message, context)
	for _, handler := range l.handlers {
		handler.Handle(messageObj)
	}
}

func (l *Logger) ExportBufferedMessages() {
	for _, handler := range l.handlers {
		if bufferedHandler, handlerHasBuffer := handler.(BufferedHandlerInterface); handlerHasBuffer {
			bufferedHandler.RunExport()
		}
	}
}

func (l *Logger) Emergency(message string, context *MessageContext) {
	l.Log(EMERGENCY, message, context)
}

func (l *Logger) Critical(message string, context *MessageContext) {
	l.Log(CRITICAL, message, context)
}

func (l *Logger) Error(message string, context *MessageContext) {
	l.Log(ERROR, message, context)
}

func (l *Logger) Warning(message string, context *MessageContext) {
	l.Log(WARNING, message, context)
}

func (l *Logger) Notice(message string, context *MessageContext) {
	l.Log(NOTICE, message, context)
}

func (l *Logger) Info(message string, context *MessageContext) {
	l.Log(INFO, message, context)
}

func (l *Logger) Debug(message string, context *MessageContext) {
	l.Log(DEBUG, message, context)
}

func (l *Logger) SetHandlers(handlers []HandlerInterface) *Logger {
	l.handlers = handlers
	return l
}