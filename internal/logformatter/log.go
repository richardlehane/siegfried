package logformatter

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogWriter struct {
	Appname string
	UTC     bool
}

const logTimeFormat = "2006-01-02 15:04:05"

// Write enables us to format a logging prefix for the application. The
// text will appear before the log message output by the caller.
//
// e.g.
//
//	`// 2023-11-27 11:36:57 ERROR :: golang-app:100:main() :: this is an error message, ...some diagnosis`
func (lw *LogWriter) Write(logString []byte) (int, error) {
	logTime := time.Now().UTC().Format(logTimeFormat)
	if !lw.UTC {
		logTime = time.Now().Format(logTimeFormat)
	}
	return fmt.Fprintf(os.Stderr, "%s :: %s :: %s",
		logTime,
		lw.Appname,
		string(logString),
	)
}

func init() {
	// Configure logging to use a custom log writer with sensible defaults.
	log.SetFlags(0 | log.Lshortfile | log.LUTC)
	log.SetOutput(new(LogWriter))
}
