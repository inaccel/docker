package grep

import (
	"bytes"
	"io"
	"os"
	"regexp"
)

var (
	color = []byte("\033[01;31m")
	reset = []byte("\033[00m")
)

func init() {
	if env, ok := os.LookupEnv("GREP_COLOR"); ok {
		color = []byte("\033[" + env + "m")
	}
}

type Grep struct {
	ere *regexp.Regexp
}

type internal struct {
	*Grep

	writer io.Writer
	color  bool

	buffer []byte
}

func Compile(pattern string) (*Grep, error) {
	ere, err := regexp.CompilePOSIX(pattern)
	if err != nil {
		return nil, err
	}

	return &Grep{
		ere: ere,
	}, nil
}

func MustCompile(pattern string) *Grep {
	ere := regexp.MustCompilePOSIX(pattern)

	return &Grep{
		ere: ere,
	}
}

func (grep *Grep) WriteCloser(writer io.Writer, color bool) io.WriteCloser {
	return &internal{
		Grep:   grep,
		writer: writer,
		color:  color,
	}
}

func (grep *internal) flush(line []byte) error {
	if grep.ere.Match(line) {
		if grep.color {
			line = grep.ere.ReplaceAllFunc(line, func(match []byte) []byte {
				return append(append(color, match...), reset...)
			})
		}

		if _, err := grep.writer.Write(line); err != nil {
			return err
		}
	}

	return nil
}

func (grep *internal) Close() error {
	return grep.flush(grep.buffer)
}

func (grep *internal) Write(p []byte) (int, error) {
	var offset int
	for {
		index := bytes.IndexByte(p[offset:], '\n')
		if index < 0 {
			grep.buffer = append(grep.buffer, p[offset:]...)

			return len(p), nil
		}

		var line []byte
		if len(grep.buffer) > 0 {
			line = append(grep.buffer, p[offset:offset+index+1]...)

			grep.buffer = nil
		} else {
			line = p[offset : offset+index+1]
		}

		offset += index + 1

		if err := grep.flush(line); err != nil {
			return -1, err
		}
	}
}
