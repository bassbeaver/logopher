package logopher

import (
	"io"
	"gopkg.in/mgo.v2"
	"github.com/bassbeaver/logopher/dateformat"
	"sync"
)

type ExporterInterface interface {
	ExportMessage(message *Message) error
}

type HandlerInterface interface {
	Handle(message *Message) error
	SetAcceptLevels(acceptLevels *map[string]bool)
	SetSkipLevels(skipLevels *map[string]bool)
	AcceptMessage(message *Message) bool
}

type BufferedHandlerInterface interface {
	ExporterInterface
	HandlerInterface
	RunExport() error
}

// --------------------------------------------

type AbstractHandler struct {
	acceptLevels *map[string]bool
	skipLevels *map[string]bool
}

func (h *AbstractHandler) SetAcceptLevels(acceptLevels *map[string]bool) {
	h.acceptLevels = acceptLevels
}

func (h *AbstractHandler) SetSkipLevels(skipLevels *map[string]bool) {
	h.skipLevels = skipLevels
}

func (h *AbstractHandler) AcceptMessage(message *Message) bool {
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
	buffer struct{
		data         []*Message
		mutex sync.Mutex
		size         int
	}
}

func (h *BufferedHandler) initBuffer(size int) {
	h.buffer.data = make([]*Message, 0)
	h.buffer.size = size
}

func (h *BufferedHandler) RunExport() error {
	h.buffer.mutex.Lock()
	defer h.buffer.mutex.Unlock()

	for _, message := range h.buffer.data {
		err := h.ExportMessage(message)
		if nil != err {
			return err
		}
	}
	h.buffer.data = make([]*Message, 0)

	return nil
}

func (h *BufferedHandler) Handle(message *Message) error {
	if !h.AcceptMessage(message) {
		return nil
	}

	var err error

	h.buffer.mutex.Lock()
	h.buffer.data = append(h.buffer.data, message)
	h.buffer.mutex.Unlock()

	if len(h.buffer.data) >= h.buffer.size {
		err = h.RunExport()
	}

	return err
}

// --------------------------------------------

type StreamHandler struct {
	BufferedHandler
	formatter FormatterInterface
	writer io.Writer
}

func (h *StreamHandler) ExportMessage(message *Message) error {
	formattedMessage := h.formatter.Format(message) + "\n"
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

func (h *MongoHandler) ExportMessage(message *Message) error {
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
		},
		formatter: formatter,
		writer: writer,
	}
	streamHandler.initBuffer(bufferSize)
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
		},
		dbName: dbName,
		collectionName: collectionName,
		mongodbSession: mongodbSession,
	}
	mongoHandler.initBuffer(bufferSize)
	mongoHandler.BufferedHandler.ExporterInterface = &mongoHandler

	return &mongoHandler
}