package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/rayakame/sqlc-gen-better-python/internal/utils"
)

var (
	loggingInstance *Logger
	loggingOnce     sync.Once
)

type Logger struct {
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
		loggingInstance = utils.ToPtr(Logger{})
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
	joined := strings.Join(logger.messages, ",\n")
	if joined != "" {
		joined += "\n"
	}

	return "log.json", []byte("[\n" + joined + "]")
}

func (logger *Logger) log(data string) {
	logger.messages = append(logger.messages, data)
}
