package protocol

type ValueType byte

const (
	TypeSimpleString ValueType = '+'
	TypeError        ValueType = '-'
	TypeInteger      ValueType = ':'
	TypeBulkString   ValueType = '$'
	TypeArray        ValueType = '*'
)

type Value struct {
	Type  ValueType
	Str   string
	Int   int64
	Elems []Value
	Nil   bool
}
