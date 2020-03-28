package wiki

import (
	"io/ioutil"
	"log"
	"os"
)

// Log logs info for a wiki.
func (w *Wiki) Log(i ...interface{}) {
	w.logger().Println(i...)
	log.Println(i...)
}

// Debug logs debug info for a wiki.
func (w *Wiki) Debug(i ...interface{}) {
	w.logger().Println(i...)
	log.Println(i...)
}

// Logf logs info for a wiki.
func (w *Wiki) Logf(format string, i ...interface{}) {
	w.logger().Printf(format+"\n", i...)
	log.Printf(format+"\n", i...)
}

// Debugf logs debug info for a wiki.
func (w *Wiki) Debugf(format string, i ...interface{}) {
	w.logger().Printf(format+"\n", i...)
	log.Printf(format+"\n", i...)
}

func (w *Wiki) logger() *log.Logger {
	// consider: if wiki is ever destoryed, need to close this
	f, err := os.OpenFile(w.Dir("cache", "wiki.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return log.New(ioutil.Discard, "", log.LstdFlags)
	}
	w._logger = log.New(f, "", log.LstdFlags)
	return w._logger
}
