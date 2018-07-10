package logopher

import (
	"testing"
	"bytes"
	"math"
	"math/rand"
	"fmt"
	"time"
	"bufio"
	"sync"
	"regexp"
	"strings"
	"strconv"
)

func TestConcurrentExport(t *testing.T) {
	var bytesBuffer bytes.Buffer
	bufferHandler := CreateStreamHandler(&bytesBuffer, &SimpleFormatter{}, nil, nil, 10)

	logger := Logger{}
	logger.SetHandlers([]HandlerInterface{bufferHandler})

	var w = new(sync.WaitGroup)
	var iterations = 600
	for i:=1; i <= iterations; i++ {
		go func(waitGroup *sync.WaitGroup, goroutineNum int, loggerObj *Logger) {
			randSleep := int( math.Floor( 200 + ( 2 * rand.Float64() ) ) )

			loggerObj.Info(fmt.Sprintf("Thread %d starting to wait %d", goroutineNum, randSleep), nil)

			time.Sleep( time.Duration(randSleep) * time.Millisecond )

			loggerObj.Info(fmt.Sprintf("Thread %d waited %d", goroutineNum, randSleep), nil)

			waitGroup.Done()
		}(w, i, &logger)
		w.Add(1)
	}
	w.Wait()

	logger.ExportBufferedMessages()

	// Count number of exported lines and compare with total number of lines, also count number of log entries for each thread
	var threadsEntriesCount = make(map[int]int, 0)
	re := regexp.MustCompile(`Thread \d+`)
	scanner := bufio.NewScanner(&bytesBuffer)
	linesCount := 0
	for scanner.Scan() {
		match := re.FindStringSubmatch(scanner.Text())[0]
		threadNum, err := strconv.Atoi(strings.Replace(match, "Thread ", "", 1))
		if err != nil {
			panic(err)
		}

		if _, threadCounted := threadsEntriesCount[threadNum]; threadCounted {
			threadsEntriesCount[threadNum]++
		} else {
			threadsEntriesCount[threadNum] = 1
		}

		linesCount++
	}

	loggedLines := iterations*2
	if loggedLines > linesCount {
		t.Errorf("Logged lines: %d. Exported lines: %d", loggedLines, linesCount)
	}

	// Count number of log entries for each thread
	for threadId, threadCounted := range threadsEntriesCount {
		if threadCounted != 2 {
			t.Errorf("Thread %d was exported %d times", threadId, threadCounted)
		}
	}
}
