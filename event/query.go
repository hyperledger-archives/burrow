package event

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/tendermint/tmlibs/pubsub"
	"github.com/tendermint/tmlibs/pubsub/query"
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
type QueryString string

func (qs QueryString) Query() (pubsub.Query, error) {
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
type QueryBuilder struct {
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
func NewQueryBuilder(queries ...string) *QueryBuilder {
	qb := new(QueryBuilder)
	qb.queryString = qb.and(stringIterator(queries...))
	return qb
}

func (qb *QueryBuilder) String() string {
	return qb.queryString
}

func (qb *QueryBuilder) Query() (pubsub.Query, error) {
	if qb.error != nil {
		return nil, qb.error
	}
	if isEmpty(qb.queryString) {
		return query.Empty{}, nil
	}
	return query.New(qb.String())
}

// Creates the conjunction of QueryBuilder and rightQuery
func (qb *QueryBuilder) And(queryBuilders ...*QueryBuilder) *QueryBuilder {
	return NewQueryBuilder(qb.and(queryBuilderIterator(queryBuilders...)))
}

// Creates the conjunction of QueryBuilder and tag = operand
func (qb *QueryBuilder) AndEquals(tag string, operand interface{}) *QueryBuilder {
	qb.condition.Tag = tag
	qb.condition.Op = equalString
	qb.condition.Operand = qb.operand(operand)
	return NewQueryBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *QueryBuilder) AndGreaterThanOrEqual(tag string, operand interface{}) *QueryBuilder {
	qb.condition.Tag = tag
	qb.condition.Op = greaterOrEqualString
	qb.condition.Operand = qb.operand(operand)
	return NewQueryBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *QueryBuilder) AndLessThanOrEqual(tag string, operand interface{}) *QueryBuilder {
	qb.condition.Tag = tag
	qb.condition.Op = lessOrEqualString
	qb.condition.Operand = qb.operand(operand)
	return NewQueryBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *QueryBuilder) AndStrictlyGreaterThan(tag string, operand interface{}) *QueryBuilder {
	qb.condition.Tag = tag
	qb.condition.Op = greaterThanString
	qb.condition.Operand = qb.operand(operand)
	return NewQueryBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *QueryBuilder) AndStrictlyLessThan(tag string, operand interface{}) *QueryBuilder {
	qb.condition.Tag = tag
	qb.condition.Op = lessThanString
	qb.condition.Operand = qb.operand(operand)
	return NewQueryBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *QueryBuilder) AndContains(tag string, operand interface{}) *QueryBuilder {
	qb.condition.Tag = tag
	qb.condition.Op = containsString
	qb.condition.Operand = qb.operand(operand)
	return NewQueryBuilder(qb.and(stringIterator(qb.conditionString())))
}

func (qb *QueryBuilder) and(queryIterator func(func(string))) string {
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

func (qb *QueryBuilder) operand(operand interface{}) string {
	defer qb.Buffer.Reset()
	switch oper := operand.(type) {
	case string:
		qb.Buffer.WriteByte('\'')
		qb.Buffer.WriteString(oper)
		qb.Buffer.WriteByte('\'')
		return qb.Buffer.String()
	case fmt.Stringer:
		return qb.operand(oper.String())
	case bool:
		if oper {
			return trueString
		}
		return falseString
	case int:
		return strconv.FormatInt(int64(oper), 10)
	case int64:
		return strconv.FormatInt(oper, 10)
	case uint:
		return strconv.FormatUint(uint64(oper), 10)
	case uint64:
		return strconv.FormatUint(oper, 10)
	case float32:
		return strconv.FormatFloat(float64(oper), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(oper), 'f', -1, 64)
	case time.Time:
		qb.Buffer.WriteString(timeString)
		qb.Buffer.WriteByte(' ')
		qb.Buffer.WriteString(oper.Format(time.RFC3339))
		return qb.Buffer.String()
	default:
		return fmt.Sprintf("%v", oper)
	}
}

func (qb *QueryBuilder) conditionString() string {
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

func queryBuilderIterator(qbs ...*QueryBuilder) func(func(string)) {
	return func(callback func(string)) {
		for _, qb := range qbs {
			callback(qb.String())
		}
	}
}
