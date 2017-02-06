package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mgutz/ansi"
)

type ResultSet struct {
	name     string
	id       string
	kind     string
	at       time.Time
	vals     map[string]bool
	status   Status
	err      error
	output   string
	children []ResultSet
}

func (rs ResultSet) PrettyPrint(level int, ts bool, vals bool) string {
	ident := strings.Repeat("   ", level)
	out := rs.status.Colorize(fmt.Sprintf("%s%s %s is %v", ident, rs.kind, rs.name, rs.status))

	ident = strings.Repeat("   ", level+1)
	if ts {
		timestamp := rs.at.Format("2006-01-02 15:04:05")
		out += fmt.Sprintf(" (%s)", timestamp)
	}
	if rs.err != nil {
		out += fmt.Sprintf("\n%sError occured: %s", ident, rs.err.Error())
	}
	if rs.status == StatusNOK && rs.output != "" {
		out += fmt.Sprintf("\n%sMessage from Monitoring: %s", ident, rs.output)
	}
	if vals && len(rs.vals) > 0 {
		out += fmt.Sprintf("\n%sValues:", ident)
		for key, value := range rs.vals {
			out += " "
			if !value {
				out += ansi.Color(key, "magenta+s")
			} else {
				out += ansi.Color(key, "green")
			}
		}
	}
	out += "\n"
	for _, childRs := range rs.children {
		out += childRs.PrettyPrint(level+1, ts, vals)
	}
	return out
}

func (rs ResultSet) AsInflux(parentTags map[string]string, saveOK []string) []Point {
	var out []Point

	tags := make(map[string]string)
	for k, v := range parentTags {
		tags[k] = v
	}
	tags[rs.kind] = rs.id

	if rs.status != StatusOK || stringInSlice(rs.kind, saveOK) {
		fields := map[string]interface{}{
			"status": rs.status.toInt(),
		}
		for key, value := range rs.vals {
			fields[key] = value
		}
		if rs.output != "" {
			fields["output"] = fmt.Sprintf("Output: %s", rs.output)
		}
		if rs.err != nil {
			fields["err"] = fmt.Sprintf("Error: %s", rs.err.Error())
		}
		pt := Point{
			Timestamp: rs.at,
			Series:    rs.kind,
			Tags:      tags,
			Fields:    fields,
		}
		out = append(out, pt)
	}

	for _, childRs := range rs.children {
		out = append(out, childRs.AsInflux(tags, saveOK)...)
	}
	return out
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToUpper(b) == strings.ToUpper(a) {
			return true
		}
	}
	return false
}
