package query

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/tendermint/tendermint/libs/pubsub"
	"github.com/tendermint/tendermint/libs/pubsub/query"
)

const (
	// Operators
	equalString          = "="
	greaterThanString    = ">"
	lessThanString       = "<"
	greaterOrEqualString = ">="
	lessOrEqualString    = "<="
	containsString       = "CONTAINS"
	andString            = "AND"

	// Values
	trueString  = "true"
	falseString = "false"
	emptyString = "empty"
	timeString  = "TIME"
	dateString  = "DATE"
)

type Queryable interface {
	Query() (pubsub.Query, error)
}

// A yet-to-parsed query
type String string

func (qs String) Query() (pubsub.Query, error) {
	if isEmpty(string(qs)) {
		return query.Empty{}, nil
	}
	return query.New(string(qs))
}

func MatchAllQueryable() Queryable {
	return WrapQuery(query.Empty{})
}

// A pre-parsed query
type Query struct {
	query pubsub.Query
}

func WrapQuery(qry pubsub.Query) Query {
	return Query{qry}
}

func (q Query) Query() (pubsub.Query, error) {
	return q.query, nil
}

// A fluent query builder
type Builder struct {
	queryString string
	condition
	// reusable buffer for building queryString
	bytes.Buffer
	error
}

// Templates
type condition struct {
	Tag     string
	Op      string
	Operand string
}

var conditionTemplate = template.Must(template.New("condition").Parse("{{.Tag}} {{.Op}} {{.Operand}}"))

// Creates a new query builder with a base query that is the conjunction of all queries passed
func NewBuilder(queries ...string) *Builder {
	qb := new(Builder)
	qb.queryString = qb.and(stringIterator(queries...))
	return qb
}

func (qb *Builder) String() string {
	return qb.queryString
}

func (qb *Builder) Query() (pubsub.Query, error) {
	if qb.error != nil {
		return nil, qb.error
	}
	if isEmpty(qb.queryString) {
		return query.Empty{}, nil
	}
	return query.New(qb.String())
}

// Creates the conjunction of Builder and rightQuery
func (qb *Builder) And(queryBuilders ...*Builder) *Builder {
	return NewBuilder(qb.and(queryBuilderIterator(queryBuilders...)))
}

// Creates the conjunction of Builder and tag = operand
func (qb *Builder) AndEquals(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = equalString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndGreaterThanOrEqual(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = greaterOrEqualString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndLessThanOrEqual(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = lessOrEqualString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndStrictlyGreaterThan(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = greaterThanString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndStrictlyLessThan(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = lessThanString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) AndContains(tag string, operand interface{}) *Builder {
	qb.condition.Tag = tag
	qb.condition.Op = containsString
	qb.condition.Operand = operandString(operand)
	return NewBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *Builder) and(queryIterator func(func(string))) string {
	defer qb.Buffer.Reset()
	qb.Buffer.WriteString(qb.queryString)
	queryIterator(func(q string) {
		if !isEmpty(q) {
			if qb.Buffer.Len() > 0 {
				qb.Buffer.WriteByte(' ')
				qb.Buffer.WriteString(andString)
				qb.Buffer.WriteByte(' ')
			}
			qb.Buffer.WriteString(q)
		}
	})
	return qb.Buffer.String()
}

func operandString(value interface{}) string {
	buf := new(bytes.Buffer)
	switch v := value.(type) {
	case string:
		buf.WriteByte('\'')
		buf.WriteString(v)
		buf.WriteByte('\'')
		return buf.String()
	case fmt.Stringer:
		return operandString(v.String())
	default:
		return StringFromValue(v)
	}
}

func StringFromValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case bool:
		if v {
			return trueString
		}
		return falseString
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case time.Time:
		return timeString + " " + v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (qb *Builder) conditionString() string {
	defer qb.Buffer.Reset()
	err := conditionTemplate.Execute(&qb.Buffer, qb.condition)
	if err != nil && qb.error == nil {
		qb.error = err
	}
	return qb.Buffer.String()
}

func isEmpty(queryString string) bool {
	return queryString == "" || queryString == emptyString
}

// Iterators over some strings
func stringIterator(strs ...string) func(func(string)) {
	return func(callback func(string)) {
		for _, s := range strs {
			callback(s)
		}
	}
}

func queryBuilderIterator(qbs ...*Builder) func(func(string)) {
	return func(callback func(string)) {
		for _, qb := range qbs {
			callback(qb.String())
		}
	}
}
