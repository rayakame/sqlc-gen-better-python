package log

import (
	"encoding/json"
	"fmt"
)

type Logger struct {
	messages []string
}

func (logger *Logger) Log(message string) {
	logger.messages = append(logger.messages, message)
}
func (logger *Logger) LogByte(message []byte) {
	logger.messages = append(logger.messages, string(message))
}

func (logger *Logger) LogAny(message any) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.Log(fmt.Sprintf("Error while trying to log any: %e", err))
	} else {
		logger.LogByte(jsonData)
	}
}

func (logger *Logger) Print() (string, []byte) {
	var loggedMessages string
	for _, message := range logger.messages {
		loggedMessages += message + "\n"
	}
	return "log.txt", []byte(loggedMessages)
}
