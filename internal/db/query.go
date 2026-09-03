package db

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	query strings.Builder
	args  []any
}

func NewQueryBuilder(initialQuery string, args ...any) *QueryBuilder {
	query := strings.Builder{}
	query.WriteString(initialQuery)
	return &QueryBuilder{query: query, args: args}
}

func (qb *QueryBuilder) And(condition string, arg any) *QueryBuilder {
	if len(condition) > 0 {
		qb.query.WriteString(" AND ")
		qb.query.WriteString(condition)
		qb.addArg(arg)
	}
	return qb
}

func (qb *QueryBuilder) Or(condition string, arg any) *QueryBuilder {
	if len(condition) > 0 {
		qb.query.WriteString(" OR ")
		qb.query.WriteString(condition)
		qb.addArg(arg)
	}
	return qb
}

func (qb *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	if len(orderBy) > 0 {
		qb.query.WriteString(" ORDER BY ")
		qb.query.WriteString(orderBy)
	}
	return qb
}

func (qb *QueryBuilder) addArg(arg any) {
	qb.args = append(qb.args, arg)
	qb.query.WriteString(fmt.Sprintf("=$%d", len(qb.args)))
}

func (qb *QueryBuilder) Build() (string, []any) {
	return qb.query.String(), qb.args
}
