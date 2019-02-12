package loge

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Various log output modes
const (
	OutputConsole             uint32 = 1  // OutputConsole outputs to the stderr
	OutputFile                uint32 = 2  // OutputFile adds a file output
	OutputFileRotate          uint32 = 4  // OutputFileRotate adds an automatic file rotation based on current date
	OutputIncludeLine         uint32 = 8  // Include file and line into the output
	OutputConsoleInJSONFormat uint32 = 16 // Switch console output to JSON serialized format
)

// Various selectable log levels
const (
	LogLevelInfo  uint32 = 1
	LogLevelDebug uint32 = 2
)

// Configuration defines the logger startup configuration
type Configuration struct {
	Mode                     uint32        // work mode
	Path                     string        // output path for the file mode
	Filename                 string        // log file name (ignored if rotation is enabled)
	TransactionSize          int           // transaction size limit in bytes (default 10KB)
	TransactionTimeout       time.Duration // transaction length limit (default 3 seconds)
	ConsoleOutput            io.Writer     // output writer for console (default os.Stderr)
	BacklogExpirationTimeout time.Duration // transaction backlog expiration timeout (default is time.Hour)
	LogLevels                uint32        // selectable log levels
}

var std *logger

func init() {
	std = newLogger(
		Configuration{
			Mode:          OutputConsole,
			ConsoleOutput: os.Stderr,
		})
}

const (
	defaultTransactionSize   = 10 * 1024
	defaultTransactionLength = time.Second * 3
	defaultBacklogTimeout    = time.Hour
)

type logger struct {
	configuration        Configuration
	writeTimestampBuffer []byte
	buffer               *buffer

	customTimestampBuffer []byte
	customTimestampLock   sync.Mutex
}

// Init initializes the library and returns the shutdown handler to defer
func Init(c Configuration) func() {
	std = newLogger(c)
	return std.shutdown
}

func newLogger(c Configuration) *logger {
	l := &logger{
		configuration: c,
	}

	flag := 0
	if (c.Mode & OutputIncludeLine) != 0 {
		flag |= log.Lshortfile
	}

	if (c.Mode & OutputFile) != 0 {
		validPath := false

		if fileInfo, err := os.Stat(c.Path); !os.IsNotExist(err) {
			if fileInfo.IsDir() {
				validPath = true
			}
		}

		if !validPath {
			l.configuration.Mode = l.configuration.Mode & (^OutputFile)
		}
	}

	if l.configuration.TransactionSize == 0 {
		l.configuration.TransactionSize = defaultTransactionSize
	}

	if l.configuration.TransactionTimeout == 0 {
		l.configuration.TransactionTimeout = defaultTransactionLength
	}

	if l.configuration.ConsoleOutput == nil {
		l.configuration.ConsoleOutput = os.Stderr
	}

	if l.configuration.BacklogExpirationTimeout == 0 {
		l.configuration.BacklogExpirationTimeout = defaultBacklogTimeout
	}

	if (l.configuration.Mode & OutputFile) != 0 {
		l.buffer = newBuffer(l)
	}

	log.SetFlags(flag)
	log.SetOutput(l)

	return l
}

func (l *logger) shutdown() {
	if l.buffer != nil {
		l.buffer.shutdown()
	}
}

func (l *logger) Write(d []byte) (int, error) {
	if ((l.configuration.Mode & OutputFile) != 0) || ((l.configuration.Mode & OutputConsole) != 0) {
		t := time.Now()
		dumpTimeToBuffer(&l.writeTimestampBuffer, t) // don't have to lock this buf here because Write events are serialized
		l.write(
			NewBufferElement(t, l.writeTimestampBuffer, d),
		)
	}

	return len(d), nil
}

func (l *logger) write(be *BufferElement) {
	if (l.configuration.Mode & OutputConsole) != 0 {
		if (l.configuration.Mode & OutputConsoleInJSONFormat) != 0 {
			json, err := be.Marshal()
			if err == nil {
				l.configuration.ConsoleOutput.Write(json)
				l.configuration.ConsoleOutput.Write([]byte("\n"))
			}
		} else {
			l.configuration.ConsoleOutput.Write(be.Timestring[:])
			l.configuration.ConsoleOutput.Write([]byte(be.Message))
			l.configuration.ConsoleOutput.Write([]byte("\n"))
		}
	}

	if (l.configuration.Mode & OutputFile) != 0 {
		l.buffer.write(
			be,
		)
	}
}

func (l *logger) writeLevel(level uint32, message string) {
	if ((l.configuration.Mode & OutputFile) != 0) || ((l.configuration.Mode & OutputConsole) != 0) {
		if (l.configuration.LogLevels & level) != 0 {
			l.customTimestampLock.Lock()
			defer l.customTimestampLock.Unlock()
			t := time.Now()
			dumpTimeToBuffer(&l.customTimestampBuffer, t)
			be := NewBufferElement(t, l.writeTimestampBuffer, []byte(message))
			switch level {
			case LogLevelInfo:
				be.Level = "info"
			case LogLevelDebug:
				be.Level = "debug"
			}
			l.write(be)
		}
	}
}

// Infof creates creates a new "info" log entry
func Infof(format string, v ...interface{}) {
	std.writeLevel(LogLevelInfo, fmt.Sprintf(format, v...))
}

// Infoln creates creates a new "info" log entry
func Infoln(v ...interface{}) {
	std.writeLevel(LogLevelInfo, fmt.Sprintln(v...))
}

// Debugf creates creates a new "debug" log entry
func Debugf(format string, v ...interface{}) {
	std.writeLevel(LogLevelDebug, fmt.Sprintf(format, v...))
}

// Debugln creates creates a new "debug" log entry
func Debugln(v ...interface{}) {
	std.writeLevel(LogLevelDebug, fmt.Sprintln(v...))
}