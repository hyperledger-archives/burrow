package events

type Type int8

// Execution event types
const (
	TypeCall          = Type(0x00)
	TypeLog           = Type(0x01)
	TypeAccountInput  = Type(0x02)
	TypeAccountOutput = Type(0x03)
)

var nameFromType = map[Type]string{
	TypeCall:          "CallEvent",
	TypeLog:           "LogEvent",
	TypeAccountInput:  "AccountInputEvent",
	TypeAccountOutput: "AccountOutputEvent",
}

var typeFromName = make(map[string]Type)

func init() {
	for t, n := range nameFromType {
		typeFromName[n] = t
	}
}

func EventTypeFromString(name string) Type {
	return typeFromName[name]
}

func (typ Type) String() string {
	name, ok := nameFromType[typ]
	if ok {
		return name
	}
	return "UnknownTx"
}

func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.String()), nil
}

func (typ *Type) UnmarshalText(data []byte) error {
	*typ = EventTypeFromString(string(data))
	return nil
}
