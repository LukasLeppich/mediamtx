package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type destinationFile struct {
	structured bool
	file       io.WriteCloser
	buf        bytes.Buffer
}

func newDestinationFile(structured bool, filePath string, limitMB int) (destination, error) {
	var f io.WriteCloser
	var err error
	if limitMB <= 0 {
		f, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
	} else {
		f = &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    limitMB,
			MaxBackups: 0,
		}
	}

	return &destinationFile{
		structured: structured,
		file:       f,
	}, nil
}

func (d *destinationFile) log(t time.Time, level Level, format string, args ...any) {
	d.buf.Reset()

	if d.structured {
		d.buf.WriteString(`{"timestamp":"`)
		d.buf.WriteString(t.Format(time.RFC3339Nano))
		d.buf.WriteString(`","level":"`)
		writeLevel(&d.buf, level, false)
		d.buf.WriteString(`","message":`)
		d.buf.WriteString(strconv.Quote(fmt.Sprintf(format, args...)))
		d.buf.WriteString(`}`)
		d.buf.WriteByte('\n')
	} else {
		writePlainTime(&d.buf, t, false)
		writeLevel(&d.buf, level, false)
		d.buf.WriteByte(' ')
		fmt.Fprintf(&d.buf, format, args...)
		d.buf.WriteByte('\n')
	}

	d.file.Write(d.buf.Bytes()) //nolint:errcheck
}

func (d *destinationFile) close() {
	d.file.Close()
}
