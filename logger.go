package httpproxy

type Logger interface {
	Printf(format string, v ...interface{})
}
