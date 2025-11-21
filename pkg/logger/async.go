package logger

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type LogEntry struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Time    time.Time              `json:"time"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

type AsyncLogger struct {
	ch     chan LogEntry
	writer *os.File
}

func NewAsyncLogger(filePath string, buffer int) (*AsyncLogger, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &AsyncLogger{
		ch:     make(chan LogEntry, buffer),
		writer: f,
	}, nil
}

func (l *AsyncLogger) Start() {
	go func() {
		for entry := range l.ch {
			data, _ := json.Marshal(entry)
			_, err := l.writer.Write(append(data, '\n'))
			if err != nil {
				log.Error().Err(err).Msg("async logger write failed")
			}
		}
	}()
}

func (l *AsyncLogger) Stop() {
	close(l.ch)
	_ = l.writer.Close()
}

type EventBuilder struct {
	logger *AsyncLogger
	level  string
	fields map[string]interface{}
}

func (l *AsyncLogger) Error() Event {
	return &EventBuilder{
		logger: l,
		level:  "ERROR",
		fields: make(map[string]interface{}),
	}
}

func (l *AsyncLogger) Warn() Event {
	return &EventBuilder{
		logger: l,
		level:  "WARN",
		fields: make(map[string]interface{}),
	}
}

func (l *AsyncLogger) Info() Event {
	return &EventBuilder{
		logger: l,
		level:  "INFO",
		fields: make(map[string]interface{}),
	}
}

func (e *EventBuilder) Str(key, value string) Event {
	e.fields[key] = value
	return e
}

func (e *EventBuilder) Int(key string, value int) Event {
	e.fields[key] = value
	return e
}

func (e *EventBuilder) Any(key string, value interface{}) Event {
	e.fields[key] = value
	return e
}

func (e *EventBuilder) Err(err error) Event {
	if err != nil {
		e.fields["error"] = err.Error()
	}
	return e
}

func (e *EventBuilder) Msg(msg string) {
	e.logger.enqueue(e.level, msg, e.fields)
}

func (l *AsyncLogger) enqueue(level, msg string, fields map[string]interface{}) {
	select {
	case l.ch <- LogEntry{
		Level:   level,
		Message: msg,
		Time:    time.Now(),
		Fields:  fields,
	}:
	default:
	}
}
