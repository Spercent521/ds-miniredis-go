package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRESP = errors.New("invalid resp")

type Parser interface {
	ParseArrayString(reader *bufio.Reader) ([]string, error)
}

type RESPParser struct{}

func NewRESPParser() *RESPParser {
	return &RESPParser{}
}

func (p *RESPParser) ParseArrayString(reader *bufio.Reader) ([]string, error) {
	arrayHeader, err := readLine(reader)
	if err != nil {
		return nil, err
	}
	if len(arrayHeader) < 2 || arrayHeader[0] != '*' {
		return nil, fmt.Errorf("%w: expected array", ErrInvalidRESP)
	}

	count, err := strconv.Atoi(arrayHeader[1:])
	if err != nil || count < 0 {
		return nil, fmt.Errorf("%w: invalid array len", ErrInvalidRESP)
	}

	argv := make([]string, 0, count)
	for i := 0; i < count; i++ {
		bulkHeader, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		if len(bulkHeader) < 2 || bulkHeader[0] != '$' {
			return nil, fmt.Errorf("%w: expected bulk string", ErrInvalidRESP)
		}

		bulkLen, err := strconv.Atoi(bulkHeader[1:])
		if err != nil || bulkLen < 0 {
			return nil, fmt.Errorf("%w: invalid bulk len", ErrInvalidRESP)
		}

		payload := make([]byte, bulkLen+2)
		if _, err := reader.Read(payload); err != nil {
			return nil, err
		}
		if payload[bulkLen] != '\r' || payload[bulkLen+1] != '\n' {
			return nil, fmt.Errorf("%w: invalid line terminator", ErrInvalidRESP)
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
	if !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("%w: missing crlf", ErrInvalidRESP)
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}
