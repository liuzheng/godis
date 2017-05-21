package logger

import (
	"github.com/op/go-logging"
	"os"
	"path/filepath"
)

var Log = logging.MustGetLogger("example")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format1 = logging.MustStringFormatter(
	`%{time:2006-01-02 15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x} %{message}`,
)
var format = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} %{shortfunc} > %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)
// Password is just an example type implementing the Redactor interface. Any
// time this is logged, the Redacted() function will be called.
type Password string

func (p Password) Redacted() interface{} {
	return logging.Redact(string(p))
}

func Logs(logpath string) (*logging.Logger, error) {
	if _, err := os.Stat(logpath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		filePathDir := filepath.Dir(logpath)
		if _, err = os.Stat(filePathDir); os.IsNotExist(err) {
			err = os.MkdirAll(filePathDir, 0755)
			if err != nil {
				return nil, err
			}
		}

	}

	f, err := os.OpenFile(logpath, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// For demo purposes, create two backend for os.Stderr.
	backend1 := logging.NewLogBackend(f, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Formatter := logging.NewBackendFormatter(backend1, format1)
	backend1Leveled := logging.AddModuleLevel(backend1Formatter)
	backend1Leveled.SetLevel(logging.INFO, "")



	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)
	//Log.Notice("Start logging...")
	return Log, nil
}