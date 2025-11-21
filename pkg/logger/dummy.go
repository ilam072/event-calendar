package logger

type DummyLogger struct{}

func (d *DummyLogger) Info() Event  { return &DummyEvent{} }
func (d *DummyLogger) Warn() Event  { return &DummyEvent{} }
func (d *DummyLogger) Error() Event { return &DummyEvent{} }

type DummyEvent struct{}

func (e *DummyEvent) Err(error) Event          { return e }
func (e *DummyEvent) Any(string, any) Event    { return e }
func (e *DummyEvent) Str(string, string) Event { return e }
func (e *DummyEvent) Msg(string)               {}
