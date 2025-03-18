package log

import (
	"fmt"
	"sync"
)

const (
	LvDebug = iota
	LvInfo
	LvWarn
	LvError
)

type Logger interface {
	WarnS(args ...interface{})
	ErrorS(args ...interface{})
	DebugS(args ...interface{})
	InfoS(args ...interface{})
}

var Log Logger = &DefaultLogger{}

var GetLogLevel = defaultGetLogLevel

func defaultGetLogLevel() int { return LvError }

func Enabled(lvl int) bool {
	return lvl >= GetLogLevel()
}

type DefaultLogger struct{}

func (l *DefaultLogger) WarnS(args ...interface{})  { fmt.Println(args...) }
func (l *DefaultLogger) ErrorS(args ...interface{}) { fmt.Println(args...) }
func (l *DefaultLogger) DebugS(args ...interface{}) { fmt.Println(args...) }
func (l *DefaultLogger) InfoS(args ...interface{})  { fmt.Println(args...) }

var pool sync.Pool

func init() {
	pool.New = func() interface{} {
		return make([]any, 0, 4)
	}
}

func put(v []any) {
	pool.Put(v[:0])
}

func get(msg string, args map[string]any) []any {
	v := pool.Get().([]any)
	v = append(v, "message", msg)
	for key, val := range args {
		v = append(v, key, val)
	}
	return v
}

func Error(msg string, args map[string]any) {
	v := get(msg, args)
	defer put(v)
	Log.ErrorS(v...)
}
func Warn(msg string, args map[string]any) {
	v := get(msg, args)
	defer put(v)
	Log.DebugS(v...)
}
func Debug(msg string, args map[string]any) {
	v := get(msg, args)
	defer put(v)
	Log.DebugS(v...)
}

func Info(msg string, args map[string]any) {
	v := get(msg, args)
	defer put(v)
	Log.InfoS(v...)
}
