package icinga

type flag string

const (
	FlagOK                flag = "ok"
	FlagUnknown           flag = "unknown"
	FlagWarn              flag = "warn"
	FlagCritical          flag = "critical"
	FlagScheduledDowntime flag = "scheduled_downtime"
	FlagAcknowledged      flag = "acknowledged"
	FlagFailed            flag = "failed"
)

func (f flag) String() string {
	return string(f)
}

type flags map[flag]bool

var flagDefaults = flags{
	FlagOK:                false,
	FlagUnknown:           false,
	FlagWarn:              false,
	FlagCritical:          false,
	FlagScheduledDowntime: false,
	FlagAcknowledged:      false,
	FlagFailed:            true,
}

func (f flags) ToValues() map[string]bool {
	out := make(map[string]bool)
	for k, v := range flagDefaults {
		out[k.String()] = v
	}
	return out
}
