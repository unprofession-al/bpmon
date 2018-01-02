package influx

import (
	"fmt"
	"strings"
	"time"
)

type Query interface {
	Query() string
}

type SelectQuery struct {
	fields  []string
	from    string
	where   []string
	orderBy string
	desc    bool
	limit   int
}

func NewSelectQuery() *SelectQuery {
	return &SelectQuery{}
}

func (sq *SelectQuery) Fields(fields ...string) *SelectQuery {
	sq.fields = append(sq.fields, fields...)
	return sq
}

func (sq *SelectQuery) From(from string) *SelectQuery {
	sq.from = from
	return sq
}

func (sq *SelectQuery) Between(s time.Time, e time.Time) *SelectQuery {
	sq.where = append(sq.where, fmt.Sprintf("time < %d", e.UnixNano()))
	sq.where = append(sq.where, fmt.Sprintf("time > %d", s.UnixNano()))
	return sq
}

func (sq *SelectQuery) FilterTags(tags map[string]string) *SelectQuery {
	for key, value := range tags {
		sq.where = append(sq.where, fmt.Sprintf("%s = '%s'", key, value))
	}
	return sq
}

func (sq *SelectQuery) Filter(filter string) *SelectQuery {
	sq.where = append(sq.where, filter)
	return sq
}

func (sq *SelectQuery) OrderBy(orderBy string) *SelectQuery {
	sq.orderBy = orderBy
	return sq
}

func (sq *SelectQuery) Asc() *SelectQuery {
	sq.desc = false
	return sq
}

func (sq *SelectQuery) Desc() *SelectQuery {
	sq.desc = true
	return sq
}

func (sq *SelectQuery) Limit(limit int) *SelectQuery {
	sq.limit = limit
	return sq
}

func (sq *SelectQuery) Query() string {
	fields := sq.fieldsQuery()
	where := strings.Join(sq.where, " AND ")
	order := ""
	if sq.orderBy != "" {
		order = fmt.Sprintf("ORDER BY %s ", sq.orderBy)
		if sq.desc {
			order += "DESC"
		} else {
			order += "ASC"
		}
	}
	limit := ""
	if sq.limit > 0 {
		limit = fmt.Sprintf("LIMIT %d", sq.limit)
	}

	return fmt.Sprintf("SELECT %s FROM %s WHERE %s %s %s", fields, sq.from, where, order, limit)
}

func (sq *SelectQuery) fieldsQuery() string {
	// if no fields are specified, query all
	if len(sq.fields) == 0 {
		return "*"
	}

	// if a field contains on asterix, query all
	for _, field := range sq.fields {
		if field == "*" {
			return "*"
		}
	}

	// make sure time is in the filst position
	timeFound := false
	for i, field := range sq.fields {
		if field == "time" {
			timeFound = true
			if i != 0 {
				tmp := sq.fields[0]
				sq.fields[0] = "time"
				sq.fields[i] = tmp
			}
			continue
		}
	}
	if !timeFound {
		sq.fields = append([]string{"time"}, sq.fields...)
	}

	return strings.Join(sq.fields, ", ")
}
