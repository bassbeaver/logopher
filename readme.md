# Logopher

Simple logger for Golang. Inspired by PHP Monolog library.

### Installation
```bash
go get github.com/bassbeaver/logopher
```

### Concepts
Main instances of this package is `Logger`, `Handler`, `Formatter`, `Message`. 

Each `Logger` has stack of `Handlers`. When `Message` is pushed to `Logger`
`Logger` iterates through the `Handlers` stack and every `Handler` decides if it
should process this message or skip it.  

Each `Handler` has `Formatter`. If `Handler` decides to process `Message`, `Handler`
use `Formatter` to format `Message` and after that `Handler` pushes formatted `Message`
to the log storage (file, stdout (or other stream), MongoDB, etc.).

Chain of method calls for each log record is next:
```bash
Logger.Log() -> Handler.Handle() -> Handler.AcceptMessage() -> Formatter.Format() -> Handler writes to storage
```

##### Message
Message has next fields:
* `message string` - Textual message of this log record
* `timestamp *time.Time` - Date and time when log message was created
* `level string` - Level of this log record
* `context *MessageContext` - Context of this log record. "Under the hood" `MessageContext` is `map[string]interface{}`

Logohper supports next log levels for messages:

* DEBUG
* INFO
* NOTICE
* WARNING
* ERROR
* CRITICAL
* ALERT
* EMERGENCY

##### Logger
`Logger` main method to add new log record is 

```go
func (l *Logger) Log(level, message string, context *MessageContext)
```

Also `Logger` has pack of methods to write records for each log level:

```go
func (l *Logger) Emergency(message string, context *MessageContext)
func (l *Logger) Alert(message string, context *MessageContext)
func (l *Logger) Critical(message string, context *MessageContext)
func (l *Logger) Error(message string, context *MessageContext)
func (l *Logger) Warning(message string, context *MessageContext)
func (l *Logger) Notice(message string, context *MessageContext)
```

##### Handler
Every handler must implement `HandlerInterface`:

```go
type HandlerInterface interface {
    Handle(message *Message) error
    SetAcceptLevels(acceptLevels *map[string]bool)
    SetSkipLevels(skipLevels *map[string]bool)
    AcceptMessage(message *Message) bool
}
```

Also package has `BufferedHandlerInterface` and `ExporterInterface`. `BufferedHandlerInterface` can be used for bufferizing messages 
and bulk export to the log target.

```go
type ExporterInterface interface {
    ExportMessage(message *Message) error
}
```
```go
type BufferedHandlerInterface interface {
    ExporterInterface
    HandlerInterface
    RunExport() error
}
```

Logopher package contains `BufferedHandlerInterface` implementations to write to streams (stdout, file, etc.) and MongoDB.
To instantiate handlers of that types you should use factory methods:
```go
func CreateStreamHandler(
    writer io.Writer,
    formatter FormatterInterface,
    acceptLevels *map[string]bool,
    skipLevels *map[string]bool,
    bufferSize int,
) *StreamHandler 
```

```go
func CreateMongoHandler(
    mongodbSession *mgo.Session,
    dbName         string,
    collectionName string,
    acceptLevels *map[string]bool,
    skipLevels *map[string]bool,
    bufferSize int,
) *MongoHandler
```

Also for custom `Handler` implementations you can use (embed) `AbstractHandler` struct
supplied with this package. `AbstractHandler` implements `SetAcceptLevels, SetSkipLevels, AcceptMessage` methods of `HandlerInterface`.


### Basic usage

##### Example with writing to Stdout with buffer size of 1 message
```go
import (
    "github.com/bassbeaver/logopher"
    "os"
)

simpleFormatter := logopher.SimpleFormatter{}

stdoutHandler := logopher.CreateStreamHandler(os.Stdout, &simpleFormatter, nil, nil, 1)

logger := logopher.Logger{}
logger.SetHandlers([]logopher.HandlerInterface{stdoutHandler})


logger.Info("Hello world", nil)
logger.Error("Hello world, I am error!", nil)
logger.Critical("Critical error!!!", nil)
```
Please notice, that in example above we do not call logger.ExportBufferedMessages() because
buffer size was set to 1 and messages are exported to target without buffering.


##### Example with writing JSON logs to file and MongoDB (multiple Handlers usage) with buffering.
```go
import (
    "github.com/bassbeaver/logopher"
    "os"
    "fmt"
    "gopkg.in/mgo.v2"
)

curBinDir, curBinDirError := os.Getwd()
if curBinDirError != nil {
    fmt.Println("Failed to get path of current working directory")
    panic(curBinDirError)
}


jsonFormatter := logopher.JsonFormatter{}

logfile, logfileError  := os.OpenFile(curBinDir+"/log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if logfileError != nil {
    panic(logfileError)
}
defer logfile.Close()
fileHandler := logopher.CreateStreamHandler(logfile, &jsonFormatter, nil, nil, 10)


var mgosess *mgo.Session
var mgoerr error
mgosess, mgoerr = mgo.Dial("mongodb://mongouser:mongopwd@mongohost:27017")
if mgoerr != nil {
    panic(mgoerr)
}
defer mgosess.Close()

mongoHandler := logopher.CreateMongoHandler(
    mgosess,
    "mongoDbName",
    "mongoCollectionName",
    nil,
    nil,
    10,
)

logger := logopher.Logger{}
logger.SetHandlers([]logopher.HandlerInterface{fileHandler, mongoHandler})


logger.Info("Hello world", nil)
logger.Error("Hello world, I am error!", nil)
logger.Critical("Critical error!!!", nil)

logger.ExportBufferedMessages()
```
Please notice, that in example above we call `logger.ExportBufferedMessages()` because
buffer size was set to 10 and only 3 messages was logged, so without that `logger.ExportBufferedMessages()`
call no messages would be exported to targets.


##### Example with custom buffered handler (writing to Postgres with buffering).
Example shows how to export logs into table in Postgres database with custom Handler. Logs will be exported to table `logs` wih next structure: 
```postgresql
create table logs
(
    id bigserial not null constraint logs_pk primary key,
    document jsonb not null
);
```

Example uses GORM library to provide connection with Postgres DB.

```go
import (
    "encoding/json"
    "errors"
    "github.com/bassbeaver/logopher"
    "github.com/bassbeaver/logopher/dateformat"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type PgHandler struct {
    logopher.BufferedHandler
    db *gorm.DB
}

func (h *PgHandler) ExportMessage(message *logopher.Message) error {
    jsonMessage,jsonError := json.Marshal(
        map[string]interface{}{
            "timestamp": message.Timestamp.Format(dateformat.DateTimeFormat),
            "level": message.Level,
            "message": message.Message,
            "context": message.Context,
        },
    )
    if nil != jsonError {
        return errors.New("Log Message serialization to JSON failed: "+jsonError.Error())
    }

    insertError := h.db.Exec("INSERT INTO logs(document) VALUES(?)", string(jsonMessage)).Error
    if nil != insertError {
        return errors.New("Log Message export to DB failed: "+insertError.Error())
    }

    return nil
}

func NewPgHandler(fileBufferSize int, db *gorm.DB) *PgHandler {
    h := &PgHandler{
        BufferedHandler: logopher.BufferedHandler{
            AbstractHandler: logopher.AbstractHandler{},
        },
        db: db,
    }
    h.BufferedHandler.ExporterInterface = h
    h.BufferedHandler.InitBuffer(fileBufferSize)

    return h
}


connectionString := "host=postgres port=5432 user=pguser dbname=logs password=pgpass sslmode=disable"

dbConnection, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{}, )
if nil != err {
    panic("failed to connect with Postgres: " + err.Error())
}

pgHandler := NewPgHandler(50, dbConnection)

logger := &logopher.Logger{}
logger.SetHandlers([]logopher.HandlerInterface{pgHandler})

logger.Info("Hello world", nil)
logger.Error("Hello world, I am error!", nil)
logger.Critical("Critical error!!!", nil)

logger.ExportBufferedMessages()
```