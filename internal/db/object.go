package db

type ObjectType uint8

const (
	StringType ObjectType = iota
)

type Object struct {
	Type       ObjectType
	Str        string
	ExpireAtMs int64
	LastAccess int64
}
