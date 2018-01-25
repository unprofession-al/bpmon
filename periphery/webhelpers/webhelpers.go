//go:generate esc -o static.go -pkg webhelpers -prefix static static

package webhelpers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Respond reads the 'f' url parameter ('f' stands for 'format'), formats the given data
// accordingly and sets the required content-type header. Default format is json.
func Respond(res http.ResponseWriter, req *http.Request, code int, data interface{}) {
	var err error
	var errMesg []byte
	var out []byte

	f := "json"
	format := req.URL.Query()["f"]
	if len(format) > 0 {
		f = format[0]
	}

	if f == "yaml" {
		res.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		out, err = yaml.Marshal(data)
		errMesg = []byte("--- error: failed while rendering data to yaml")
	} else {
		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		out, err = json.Marshal(data)
		errMesg = []byte("{ 'error': 'failed while rendering data to json' }")
	}

	if err != nil {
		out = errMesg
		code = http.StatusInternalServerError
	}
	res.WriteHeader(code)
	res.Write(out)
}

func GetAssetHandler(prefix string) http.Handler {
	assetFS := FS(false)
	return http.StripPrefix(prefix, http.FileServer(assetFS))
}

func GetStartEnd(req *http.Request) (start time.Time, end time.Time) {
	end = time.Now()
	start = end.AddDate(0, -1, 0)

	startStr := req.URL.Query()["start"]
	if len(startStr) > 0 {
		i, err := strconv.ParseInt(startStr[0], 10, 64)
		if err == nil {
			start = time.Unix(i, 0)
		}
	}

	endStr := req.URL.Query()["end"]
	if len(endStr) > 0 {
		i, err := strconv.ParseInt(endStr[0], 10, 64)
		if err == nil {
			end = time.Unix(i, 0)
		}
	}

	return
}
