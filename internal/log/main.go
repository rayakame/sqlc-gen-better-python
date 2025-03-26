package log

var GlobalLogger Logger

func init() {
	GlobalLogger = Logger{}
}
