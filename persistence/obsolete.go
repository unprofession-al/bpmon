package persistence

import "time"

func getInfluxTimestamp(t time.Time) int64 {
	return t.UnixNano()
}

const (
	IdentifierBusinessProcess         = "BP"
	IdentifierKeyPerformanceIndicator = "KPI"
	IdentifierService                 = "SVC"
)

func getKind(spec map[string]string) string {
	kind := "UNKNOWN"
	if _, ok := spec[IdentifierBusinessProcess]; ok {
		kind = IdentifierBusinessProcess
	}
	if _, ok := spec[IdentifierKeyPerformanceIndicator]; ok {
		kind = IdentifierKeyPerformanceIndicator
	}
	if _, ok := spec[IdentifierService]; ok {
		kind = IdentifierService
	}
	return kind
}
