package protocol

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseSimpleSetCommand(t *testing.T) {
	raw := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	p := NewRESPParser()
	argv, err := p.ParseArrayString(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(argv) != 3 || argv[0] != "SET" || argv[1] != "foo" || argv[2] != "bar" {
		t.Fatalf("unexpected argv: %v", argv)
	}
}

func TestParseGetCommand(t *testing.T) {
	raw := "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n"
	p := NewRESPParser()
	argv, err := p.ParseArrayString(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(argv) != 2 || argv[0] != "GET" || argv[1] != "foo" {
		t.Fatalf("unexpected argv: %v", argv)
	}
}

func TestParseInvalidHeader(t *testing.T) {
	raw := "wrong input\r\n"
	p := NewRESPParser()
	_, err := p.ParseArrayString(bufio.NewReader(strings.NewReader(raw)))
	if err == nil {
		t.Fatal("expected error for invalid header")
	}
}
