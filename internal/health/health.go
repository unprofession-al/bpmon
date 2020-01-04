package health

import (
	"time"

	"github.com/unprofession-al/bpmon/internal/checker"
	"github.com/unprofession-al/bpmon/internal/math"
	"github.com/unprofession-al/bpmon/internal/status"
	"github.com/unprofession-al/bpmon/internal/store"
)

type Health struct {
	StoreRequired   bool
	CheckerRequired bool
	Responsible     string
	Name            string
	ID              string
}

func New(c Config) Health {
	h := Health{
		StoreRequired:   c.StoreRequired,
		CheckerRequired: c.CheckerRequired,
		Responsible:     c.Responsible,
		Name:            c.Name,
		ID:              c.ID,
	}
	return h
}

func (h Health) Check(c checker.Checker, s store.Accessor) *store.ResultSet {
	tags := map[store.Kind]string{store.KindBusinessProcess: h.ID}
	rs := store.ResultSet{
		Name:          h.Name,
		ID:            h.ID,
		Start:         time.Now(),
		Tags:          tags,
		Children:      []*store.ResultSet{},
		Was:           status.StatusUnknown,
		StatusChanged: false,
	}

	checkerOut, checkerErr := c.Health()
	checkerKPI := getHealthKPI(checkerOut, checkerErr, h.CheckerRequired, rs, "checker")
	rs.Children = append(rs.Children, checkerKPI)

	storeOut, storeErr := s.Health()
	storeKPI := getHealthKPI(storeOut, storeErr, h.StoreRequired, rs, "store")
	rs.Children = append(rs.Children, storeKPI)

	var statusValues []bool
	statusValues = append(statusValues, checkerKPI.Status.Bool())
	statusValues = append(statusValues, storeKPI.Status.Bool())
	ok, _ := math.Calculate("AND", statusValues)
	rs.Status = status.FromBool(ok)

	return &rs
}

func getHealthKPI(out string, err error, required bool, parent store.ResultSet, kind string) *store.ResultSet {
	id := parent.ID + "_" + kind + "_kpi"
	name := parent.Name + " " + kind + " KPI"

	stat := status.StatusOK
	if required && err != nil {
		stat = status.StatusNOK
	}

	tags := parent.Tags
	tags[store.KindKeyPerformanceIndicator] = id

	rs := store.ResultSet{
		Name:          name,
		ID:            id,
		Start:         time.Now(),
		Tags:          tags,
		Children:      []*store.ResultSet{},
		Status:        stat,
		Was:           status.StatusUnknown,
		StatusChanged: false,
	}

	svc := getHealthSVC(out, err, parent, kind)

	rs.Children = append(rs.Children, svc)
	return &rs
}

func getHealthSVC(out string, err error, parent store.ResultSet, kind string) *store.ResultSet {
	id := parent.ID + "_" + kind + "_svc"
	name := parent.Name + " " + kind + " SVC"

	stat := status.StatusOK
	if err != nil {
		stat = status.StatusNOK
	}

	tags := parent.Tags
	tags[store.KindService] = id

	rs := store.ResultSet{
		Name:          name,
		ID:            id,
		Start:         time.Now(),
		Tags:          tags,
		Children:      []*store.ResultSet{},
		Status:        stat,
		Was:           status.StatusUnknown,
		StatusChanged: false,
		Output:        out,
		Err:           err,
	}

	return &rs
}
