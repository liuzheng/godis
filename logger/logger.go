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

func Logs(logpath, frontend, backend string) (*logging.Logger, error) {
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
    Backend := logging.NewLogBackend(f, "", 0)
    Frontend := logging.NewLogBackend(os.Stderr, "", 0)

    // For messages written to backend2 we want to add some additional
    // information to the output, including the used log level and the name of
    // the function.
    FrontendFormatter := logging.NewBackendFormatter(Frontend, format)
    FrontendLeveled := logging.AddModuleLevel(FrontendFormatter)
    level, _ := logging.LogLevel(frontend)
    FrontendLeveled.SetLevel(level, "")


    // Only errors and more severe messages should be sent to backend1
    BackendFormatter := logging.NewBackendFormatter(Backend, format1)
    BackendLeveled := logging.AddModuleLevel(BackendFormatter)
    level, _ = logging.LogLevel(backend)
    BackendLeveled.SetLevel(level, "")



    // Set the Backends to be used.
    logging.SetBackend(BackendLeveled, FrontendLeveled)
    //Log.Notice("Start logging...")
    return Log, nil
}