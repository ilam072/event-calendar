package logger

type Event interface {
	Err(error) Event
	Any(string, any) Event
	Str(string, string) Event
	Msg(string)
}

type Logger interface {
	Info() Event
	Warn() Event
	Error() Event
}
