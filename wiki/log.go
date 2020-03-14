package wiki

import "log"

// Log logs info for a wiki.
func (w *Wiki) Log(i ...interface{}) {
	log.Println(i...)
}

// Debug logs debug info for a wiki.
func (w *Wiki) Debug(i ...interface{}) {
	log.Println(i...)
}

// Logf logs info for a wiki.
func (w *Wiki) Logf(format string, i ...interface{}) {
	log.Printf(format+"\n", i...)
}

// Debugf logs debug info for a wiki.
func (w *Wiki) Debugf(format string, i ...interface{}) {
	log.Printf(format+"\n", i...)
}