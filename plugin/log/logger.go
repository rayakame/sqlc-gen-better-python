package log

type Logger struct {
	messages []string
}

func (logger *Logger) Log(message string) {
	logger.messages = append(logger.messages, message)
}

func (logger *Logger) Print() (string, []byte) {
	var loggedMessages string
	for _, message := range logger.messages {
		loggedMessages += message + "\n"
	}
	return "log.txt", []byte(loggedMessages)
}
