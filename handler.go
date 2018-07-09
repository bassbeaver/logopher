package logopher

import (
	"io"
	"gopkg.in/mgo.v2"
	"github.com/bassbeaver/logopher/dateformat"
)

type ExporterInterface interface {
	exportMessage(message *Message) error
}

type HandlerInterface interface {
	handle(message *Message) error
	setAcceptLevels(acceptLevels *map[string]bool)
	setSkipLevels(skipLevels *map[string]bool)
	acceptMessage(message *Message) bool
}

type BufferedHandlerInterface interface {
	ExporterInterface
	HandlerInterface
	addMessage(message *Message)
	isBufferFilled() bool
	runExport() error
}

// --------------------------------------------

type AbstractHandler struct {
	acceptLevels *map[string]bool
	skipLevels *map[string]bool
}

func (h *AbstractHandler) setAcceptLevels(acceptLevels *map[string]bool) {
	h.acceptLevels = acceptLevels
}

func (h *AbstractHandler) setSkipLevels(skipLevels *map[string]bool) {
	h.skipLevels = skipLevels
}

func (h *AbstractHandler) acceptMessage(message *Message) bool {
	if h.acceptLevels != nil && len(*h.acceptLevels) > 0 {
		_, accept := (*h.acceptLevels)[message.level]
		return accept
	}

	if h.skipLevels != nil && len(*h.skipLevels) > 0 {
		_, skip:= (*h.skipLevels)[message.level]
		return !skip
	}

	return true
}


type BufferedHandler struct {
	ExporterInterface
	AbstractHandler
	buffer []*Message
	bufferSize int
}

func (h *BufferedHandler) addMessage(message *Message) {
	h.buffer = append(h.buffer, message)
}

func (h *BufferedHandler) getBufferLength() int {
	return len(h.buffer)
}

func (h *BufferedHandler) isBufferFilled() bool {
	return len(h.buffer) >= h.bufferSize
}

func (h *BufferedHandler) runExport() error {
	var err error
	for len(h.buffer) > 0 {
		message := h.buffer[0]
		h.buffer = h.buffer[1:]
		err = h.exportMessage(message)
	}
	return err
}

func (h *BufferedHandler) handle(message *Message) error {
	if !h.acceptMessage(message) {
		return nil
	}

	var err error

	h.addMessage(message)
	if h.isBufferFilled() {
		err = h.runExport()
	}

	return err
}

// --------------------------------------------

type StreamHandler struct {
	BufferedHandler
	formatter FormatterInterface
	writer io.Writer
}

func (h *StreamHandler) exportMessage(message *Message) error {
	formattedMessage := h.formatter.format(message) + "\n"
	_, err := h.writer.Write([]byte(formattedMessage))
	return err
}

// --------------------------------------------

type MongoHandler struct {
	BufferedHandler
	dbName           string
	collectionName   string
	mongodbSession   *mgo.Session
}

func (h *MongoHandler) exportMessage(message *Message) error {
	return h.mongodbSession.DB(h.dbName).C(h.collectionName).Insert(
		map[string]interface{}{
			"timestamp": message.timestamp.Format(dateformat.DateTimeFormat),
			"level": message.level,
			"message": message.message,
			"context": message.context,
		},
	)
}

// --------------------------------------------

func CreateStreamHandler(
	writer io.Writer,
	formatter FormatterInterface,
	acceptLevels *map[string]bool,
	skipLevels *map[string]bool,
	bufferSize int,
) *StreamHandler {
	streamHandler := StreamHandler{
		BufferedHandler: BufferedHandler{
			AbstractHandler: AbstractHandler{
				acceptLevels: acceptLevels,
				skipLevels: skipLevels,
			},
			bufferSize: bufferSize,
		},
		formatter: formatter,
		writer: writer,
	}
	streamHandler.BufferedHandler.buffer = make([]*Message, 0)
	streamHandler.BufferedHandler.ExporterInterface = &streamHandler

	return &streamHandler
}

func CreateMongoHandler(
	mongodbSession *mgo.Session,
	dbName         string,
	collectionName string,
	acceptLevels *map[string]bool,
	skipLevels *map[string]bool,
	bufferSize int,
) *MongoHandler {
	mongoHandler := MongoHandler{
		BufferedHandler: BufferedHandler{
			AbstractHandler: AbstractHandler{
				acceptLevels: acceptLevels,
				skipLevels: skipLevels,
			},
			bufferSize: bufferSize,
		},
		dbName: dbName,
		collectionName: collectionName,
		mongodbSession: mongodbSession,
	}
	mongoHandler.BufferedHandler.buffer = make([]*Message, 0)
	mongoHandler.BufferedHandler.ExporterInterface = &mongoHandler

	return &mongoHandler
}