package influx

import (
	"fmt"
	"strings"
	"time"

	"github.com/unprofession-al/bpmon/internal/store"
)

type query interface {
	Query() string
}

type selectquery struct {
	fields  []string
	from    string
	where   []string
	orderBy string
	desc    bool
	limit   int
}

func newSelectQuery() *selectquery {
	return &selectquery{}
}

func (sq *selectquery) Fields(fields ...string) *selectquery {
	sq.fields = append(sq.fields, fields...)
	return sq
}

func (sq *selectquery) From(from string) *selectquery {
	sq.from = from
	return sq
}

func (sq *selectquery) Between(s time.Time, e time.Time) *selectquery {
	sq.where = append(sq.where, fmt.Sprintf("%s < %d", timefield, e.UnixNano()))
	sq.where = append(sq.where, fmt.Sprintf("%s > %d", timefield, s.UnixNano()))
	return sq
}

func (sq *selectquery) FilterTags(tags map[store.Kind]string) *selectquery {
	for key, value := range tags {
		sq.where = append(sq.where, fmt.Sprintf("%s = '%s'", key, value))
	}
	return sq
}

func (sq *selectquery) Filter(filter string) *selectquery {
	sq.where = append(sq.where, filter)
	return sq
}

func (sq *selectquery) OrderBy(orderBy string) *selectquery {
	sq.orderBy = orderBy
	return sq
}

func (sq *selectquery) Asc() *selectquery {
	sq.desc = false
	return sq
}

func (sq *selectquery) Desc() *selectquery {
	sq.desc = true
	return sq
}

func (sq *selectquery) Limit(limit int) *selectquery {
	sq.limit = limit
	return sq
}

func (sq *selectquery) Query() string {
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

func (sq *selectquery) fieldsQuery() string {
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

	// make sure time is in the first position
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
