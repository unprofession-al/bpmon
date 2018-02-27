package bpmon

import (
	"time"

	"github.com/unprofession-al/bpmon/checker"
	"github.com/unprofession-al/bpmon/status"
	"github.com/unprofession-al/bpmon/store"
)

type Heartbeat struct {
	Template        string `yaml:"template"`
	StoreRequired   bool   `yaml:"store_required"`
	CheckerRequired bool   `yaml:"checker_required"`
	Responsible     string `yaml:"responsible"`
	Name            string `yaml:"name"`
	ID              string `yaml:"id"`
}

func (hb Heartbeat) Trigger(c checker.Checker, s store.Accessor) *store.ResultSet {
	tags := map[store.Kind]string{store.KindBusinessProcess: hb.ID}
	rs := store.ResultSet{
		Name:          hb.Name,
		ID:            hb.ID,
		Start:         time.Now(),
		Tags:          tags,
		Children:      []*store.ResultSet{},
		Was:           status.StatusUnknown,
		StatusChanged: false,
	}

	checkerOut, checkerErr := c.Health()
	checkerKPI := getHeartbeatKPI(checkerOut, checkerErr, hb.CheckerRequired, rs, "checker")
	rs.Children = append(rs.Children, checkerKPI)

	storeOut, storeErr := s.Health()
	storeKPI := getHeartbeatKPI(storeOut, storeErr, hb.StoreRequired, rs, "store")
	rs.Children = append(rs.Children, storeKPI)

	var statusValues []bool
	statusValues = append(statusValues, checkerKPI.Status.Bool())
	statusValues = append(statusValues, storeKPI.Status.Bool())
	ok, _ := calculate("AND", statusValues)
	rs.Status = status.FromBool(ok)

	return &rs
}

func getHeartbeatKPI(out string, err error, required bool, parent store.ResultSet, kind string) *store.ResultSet {
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

	svc := getHeartbeatSVC(out, err, parent, kind)

	rs.Children = append(rs.Children, svc)
	return &rs
}

func getHeartbeatSVC(out string, err error, parent store.ResultSet, kind string) *store.ResultSet {
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
