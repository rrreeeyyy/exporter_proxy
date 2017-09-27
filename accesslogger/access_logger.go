package accesslogger

import (
	"fmt"
	"io"
)

type AccessLogger interface {
	Log(AccessLogRecord) error
}

type AccessLogRecord map[string]string

func New(format string, w io.Writer, fields []string) (AccessLogger, error) {
	switch format {
	case "ltsv":
		return &LTSVAccessLogger{
			w:      w,
			Fields: fields,
		}, nil
	}
	return nil, fmt.Errorf("%s is not valid format", format)
}
