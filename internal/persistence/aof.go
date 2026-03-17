package persistence

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Dispatcher is the minimal interface the AOF replayer needs.
type Dispatcher interface {
	Dispatch(argv []string) (string, error)
}

type AOF struct {
	mu   sync.Mutex
	path string
}

func NewAOF(path string) *AOF {
	return &AOF{path: path}
}

// Append writes a raw RESP-encoded command to the AOF file.
func (a *AOF) Append(rawRESP string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	f, err := os.OpenFile(a.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(rawRESP)
	return err
}

// Truncate clears the AOF file (called after FLUSHDB).
func (a *AOF) Truncate() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	err := os.Truncate(a.path, 0)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Replay reads the AOF file and re-executes every command through the dispatcher.
// Errors from individual commands are silently skipped (matching Redis behaviour).
func (a *AOF) Replay(d Dispatcher) error {
	f, err := os.Open(a.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		argv, err := parseRESPArray(reader)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		_, _ = d.Dispatch(argv)
	}
}

// parseRESPArray is a self-contained minimal RESP array reader for replay.
func parseRESPArray(reader *bufio.Reader) ([]string, error) {
	line, err := readLine(reader)
	if err != nil {
		return nil, err
	}
	if len(line) < 2 || line[0] != '*' {
		return nil, fmt.Errorf("aof: expected '*', got %q", line)
	}
	count, err := strconv.Atoi(line[1:])
	if err != nil || count < 0 {
		return nil, fmt.Errorf("aof: invalid array len")
	}
	argv := make([]string, 0, count)
	for i := 0; i < count; i++ {
		line, err = readLine(reader)
		if err != nil {
			return nil, err
		}
		if len(line) < 2 || line[0] != '$' {
			return nil, fmt.Errorf("aof: expected '$'")
		}
		bulkLen, err := strconv.Atoi(line[1:])
		if err != nil || bulkLen < 0 {
			return nil, fmt.Errorf("aof: invalid bulk len")
		}
		payload := make([]byte, bulkLen+2)
		if _, err = io.ReadFull(reader, payload); err != nil {
			return nil, err
		}
		argv = append(argv, string(payload[:bulkLen]))
	}
	return argv, nil
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}
