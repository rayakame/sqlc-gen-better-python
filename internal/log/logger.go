package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

var (
	loggingInstance *Logger
	loggingOnce     sync.Once
)

// Logger collects debug messages; the mutex only matters for parallel
// tests, the wasm plugin runs single-threaded.
type Logger struct {
	mu       sync.Mutex
	messages []string
}
type logMessage struct {
	Message string `json:"message"`
}

type errMessage struct {
	Error string `json:"error"`
}

func L() *Logger {
	loggingOnce.Do(func() {
		loggingInstance = new(Logger)
	})

	return loggingInstance
}

func (logger *Logger) LogErr(message string, err error) {
	msg := errMessage{Error: fmt.Sprintf("%s: %v", message, err)}
	logger.LogAny(msg)
}

func (logger *Logger) Log(message string) {
	msg := logMessage{Message: message}
	logger.LogAny(msg)
}

func (logger *Logger) LogAny(message any) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.log(fmt.Sprintf(`{"error": "Error while trying to log any: %v"}`, err))
	} else {
		logger.log(string(jsonData))
	}
}

func (logger *Logger) Export() (string, []byte) {
	logger.mu.Lock()
	joined := strings.Join(logger.messages, ",\n")
	logger.mu.Unlock()
	if joined != "" {
		joined += "\n"
	}

	return "log.json", []byte("[\n" + joined + "]")
}

func (logger *Logger) log(data string) {
	logger.mu.Lock()
	logger.messages = append(logger.messages, data)
	logger.mu.Unlock()
}
