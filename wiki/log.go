package wiki

import (
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Log logs info for a wiki.
func (w *Wiki) Log(i ...any) {
	w.logger().Println(i...)
	i = w.addLogPrefix(i)
	log.Println(i...)
}

// Debug logs debug info for a wiki.
func (w *Wiki) Debug(i ...any) {
	i = w.addLogPrefix(i)
	log.Println(i...)
}

// Logf logs info for a wiki.
func (w *Wiki) Logf(format string, i ...any) {
	w.logger().Printf(format+"\n", i...)
	log.Printf("["+w.Opt.Name+"] "+format+"\n", i...)
}

// Debugf logs debug info for a wiki.
func (w *Wiki) Debugf(format string, i ...any) {
	log.Printf("["+w.Opt.Name+"] "+format+"\n", i...)
}

func (w *Wiki) addLogPrefix(i []any) []any {
	return append([]any{"[" + w.Opt.Name + "]"}, i...)
}

func (w *Wiki) logger() *log.Logger {
	if w._logger != nil {
		return w._logger
	}
	// consider: if wiki is ever destoryed, need to close this
	f, err := os.OpenFile(w.Dir("cache", "wiki.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return log.New(io.Discard, "", log.LstdFlags)
	}
	w._logger = log.New(f, "", log.LstdFlags)
	return w._logger
}

// track occasional warning logging to avoid spam
var (
	occasionalWarningLastLogged = make(map[string]time.Time)
	occasionalWarningMu         sync.Mutex
)

// warnOccasionally logs a warning message at most once per 5 minutes to avoid spam
func (w *Wiki) warnOccasionally(message string) {
	occasionalWarningMu.Lock()
	defer occasionalWarningMu.Unlock()
	
	now := time.Now()
	if lastLogged, exists := occasionalWarningLastLogged[message]; !exists || now.Sub(lastLogged) >= 5*time.Minute {
		log.Println(message)
		occasionalWarningLastLogged[message] = now
	}
}
