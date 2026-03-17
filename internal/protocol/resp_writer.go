package protocol

import (
	"strconv"
	"strings"
)

func EncodeSimpleString(msg string) string {
	return "+" + msg + "\r\n"
}

func EncodeError(msg string) string {
	return "-ERR " + msg + "\r\n"
}

func EncodeInteger(n int64) string {
	return ":" + strconv.FormatInt(n, 10) + "\r\n"
}

func EncodeBulkString(msg *string) string {
	if msg == nil {
		return "$-1\r\n"
	}
	return "$" + strconv.Itoa(len(*msg)) + "\r\n" + *msg + "\r\n"
}

// EncodeArray encodes a string slice as a RESP array of bulk strings.
func EncodeArray(items []string) string {
	var sb strings.Builder
	sb.WriteString("*" + strconv.Itoa(len(items)) + "\r\n")
	for _, item := range items {
		v := item
		sb.WriteString(EncodeBulkString(&v))
	}
	return sb.String()
}
