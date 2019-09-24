package errors

import (
	"fmt"
)

type Coding struct {
	Number      uint32
	Name        string
	Description string
}

func description(description string) *Coding {
	return &Coding{Description: description}
}

func (c *Coding) Equal(other *Coding) bool {
	if c == nil {
		return false
	}
	return c.Number == other.Number
}

func (c *Coding) ErrorCode() *Coding {
	return c
}

func (c *Coding) Uint32() uint32 {
	if c == nil {
		return 0
	}
	return c.Number
}

func (c *Coding) Error() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("Error %d: %s", c.Number, c.Description)
}

func (c *Coding) ErrorMessage() string {
	if c == nil {
		return ""
	}
	return c.Description
}

func GetCode(err error) *Coding {
	exception := AsException(err)
	if exception == nil {
		return Code.None
	}
	return Code.Get(exception.GetCode())
}
