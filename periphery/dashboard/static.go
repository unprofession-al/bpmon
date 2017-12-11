// Code generated by "esc -o static.go -pkg dashboard -prefix static static"; DO NOT EDIT.

package dashboard

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		_ = f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/index.html": {
		local:   "static/index.html",
		size:    3109,
		modtime: 1513034577,
		compressed: `
H4sIAAAAAAAC/5xXX2/bNhB/bj4Fx8FFA8xStg7YFkscknlF97AlaLOHPQWUdLaupkiOPCfztx9IyY5s
S463l9a6393v/h+Y7Kv53S8Pf93/ympqlLjIwn9MSb3MOWgeBCArccEYY1kDJFlZS+eBcv7nw4fpj7yD
CEmBuL3//e4PNpe+Lox0VZa24lZFoV6x2sEi5zWR9ddpujCafLI0ZqlAWvRJaZq09P7nhWxQbfJPpjBk
rr+/uvrmh6srzhyonHvaKPA1APE+8yHGaGMh5wT/UKDknWvpPZBPS9M0RicB+B8kUaMzztK2Qllhqo24
yCp8YqWS3ue8NJokanBbHz0sGAXgWDZF/WIybDb1UBIazRQsqKd5qK3M0hzAUQWbZafCos4UmyVn3pW7
+hQ2lMfqJWdSUc6xRL2UjSlXQ3zeSi3e6sLb2fzm88fbu5tP8yyN0v3Y0gqfeokdfI7n6XBZDydqNIEm
qJBkoUKr3Br4lgW1XROL/067gjOscr5AReF3JUlOrZIl1EZV4HL+ISJsYRy7XXvU4D27d6YE75Mk4WI8
g/ZnJ3hpd3AX3Iz0UxlZYahy0Nt+7A2Ft9087DnoOd+6zHzp0FJ/ZMMyt9Rha+lx4WQDj9RYNTCREWTZ
JPckCSaCszjmOX/GiurrgFhwJWiaiMmsK54n6SjnrZWjidiVKEvbeF6LDDWBc2sb+rwXGjlx8eZNRpXI
JjlhA55kYyciS6nqIdXayWA7Ef4AkVob6rAWaqsWiM+NbmXxqF5Bnk3ylUWsQsKDjV1Z1KGefl3EI8hF
axKEE8HeNpX09axdnS1jYQPh4wv3o3ySqGSBCmnDRZIk3VZNWIcoOLFE+9ajVyLOBh8PosXP3eT/2P7C
js0jKGhAE7NSgzoIb6TohT2RZjgAcbuO79ceRWxb7Fl3XqPXM/p2VrOi5KB6g3EglfX27rXetgehNs/T
Ckii8qxYExnNmdGlwnLVovMWfPf26/c/zXbxha/LGRefa/PMOpXxeIZE543WWePVzdTrjk+Md1eDQ+qt
eDz4lcUjqygTr2Tdv1X+lbR3i3/gaP/ehaHg4rc+79hwRN1jeYu9vNKGcScyqsUDNpClVMePeXc4d4Kb
3b1sRfFQDtKlJ/xlFB5CJ5KOZab2tTTAfJzlK01pV8Dv78puH2qsYHwfZpd8n4aLj1jB6fU469j1H1Nf
/l6D20zfJ98l3yYN6uRLrMHWZsiEoLFKEurlGcrdY/aEYiPHvArUSO8uZ/0U0u4pm7Z/E1z8GwAA//8/
pjLBJQwAAA==
`,
	},

	"/main.js": {
		local:   "static/main.js",
		size:    5579,
		modtime: 1513034182,
		compressed: `
H4sIAAAAAAAC/+xYQY/bNhM9e3/FhDFg6rMjeb+gPWjXW6BNCmwDJEGTnoLAoKXRmpBMsiJlZ5u4v70g
Zdm0TSebRYskRX2iyJnhzOPj8MFFIzLDpQAuuKERvD8DAJjzHF8pLgTWNLpwU32ay6xZoDBRLAWFQYm3
jRqMYPCw4JXBejCCLtg2jv0tWQ05Mwwm0KeDh24YKyawGmxC218mhZYVxpW8oX1q5lxHscF3hkbRXqQa
YQICV/Ar3jx9pw5sR0A48aLazWI9lysaxW2WNJhjF12wBbo826AFFzkl8UzZedJtcrHnVqNpagEPaowN
akOtqWeyjmILZue1ji7O3KCSLP/xpbbz67Oz7SlspzfJ9eMbNL+8evGckoQpnsyUTogHtC3QL4QXQFuw
JzCGDx+g+xBNVfnfjcix4ALzQxj6lFzmfAlZxbSeDIQ0vOAZc9m1p3b1tK5lDas5rxAqrg0XNzBrNBeo
NahaZqg16jiOL5OcL69IB0HMlEKRv5aUOBqQKC5YjteCEl3JlX9wa8BKY+CEMtkIAxMY759CIWugdr3E
W+ACDnHZ0awRZji8OFqwvjYZLmACtN3lCh5H8AN8Px5DCuffjeF/rf/mCA/dZwom8B5miuepzWMELXFS
l82bEm/fwjq8swMWJmAWqqJkpqZ2QGyA6NihT535p1Ft6wlEaFn2dInCaFri7QmTZy+vQ8vrs/3ROsjh
TXSLxkfJDEOH2Jfm9E+yqXIQ0kCBJpsDuvQdrw6pDYMuaRgCGfwzPFdYZygMu0FtWbW+H92Lum1oWwJ6
1O31bABtmGnsFqQRpZArQY65oFfcZHPqgsWtQ7dZr9frZUwjjFMvcK+3CytLL6JdmtXIyot97/NT3uJT
7jkWrKnMKf9dUXZ+Hb64G6xhsgEszpvakWO6WQlceEtK74zetFu+9Wl5YvkkS7tf0O+w4bXXzxYVNg/N
DuF0PUkCC1Yi6KbG9kJxDUuu+azCgLFXPlzCOD4PFZMkDuE2yHSH8jg+D9iuA3OhkHNjlE6TJJM5KhQx
l0k2q1mWyUShSH7/oylvwu3dtWcLBqYb3o86RNJuMLIrtUlhy/banOra2dwudl3bfU2dW9e+i2D3doah
XrFrK2TqjLzGUTBtbONwR27R92+jpZU7AbvorjUTQpq2wU06W2/uwQQIsW/b0UoK5JJfPbpM+NXm0rh4
3GFn+AK1YQuV2g60YOZ1N0E9uKKRt33qjUfQXaz04KKFEG73FQbrulGbSlqk/ckOah6AutenvumnIPdt
dQD6toOEX8JQS7Ct762FOtQQtosfbQcWAtlYjj1vFjOs6fEGUWzkz/wd5vRxFKSbXySQKVsyXrEZr7i5
7SStbMw93ngnD+76vtu9k1Jx/VWqV1tK8K1HDX4Ff/9bf9eX3Jko7inMVjWVim/lZqn4vfRmqXh3jUrF
P09x7nHLne/xxTmhLT2FOIL7ycyjKBaNz6Kjm3FuX7ECffbyGsgmTZs5yOJYlf4nSv81ovTbFJd31s47
MeZ3slPCLBz269BpQKbb/uG+Tsm2b1017FcZ0hDkT2tzZyWx6+N6LldP0DBe7asJJ/9ye18ekqFdGJJp
3tqR7i9JnkexYjUKQ71/68hubHvE7tNu9aiL0R7Si8bQ/4/HkR9RVzzHJ3IlDv+ds1T4cqlei3Cmv6k2
z78CAAD//88DX37LFQAA
`,
	},

	"/style.css": {
		local:   "static/style.css",
		size:    1373,
		modtime: 1513033758,
		compressed: `
H4sIAAAAAAAC/+yU3YrbMBCF7/MUglJoYW3W2abZqE8zkkbWYGVk7MlfS9+9OFaSFSShN73rndGcb3Tm
GE0tJBFfVD3uzPlT/VoopZRPLNUBqQ2ilUnR/Tgfb2FoiStJvVbNa38sTk0SSdtL4fdiUQcER9zmngdy
Eqby6+eZczT2EU5aEUdirExMtptLPbiJvNPT9AzbwuZIP1Gr5m1qOylsgEEe3RnyTN+u7m1EGKYhJcy8
QwGKY+5wNcmJMVvYiSS+CD5msrza7HoaVQ17oAiGIslJFcYuPm6DzdXaD7fxbqKLfR8TiFYRvRRU6jJi
wHbtkHbsKptiGrT6tHbWrr8Xcn6mR/++Xq8K/Y47Tgd+wryZlV8WDDEeaZTb/7+HYY70PLUOaY/DY7X3
3m82Z+B/wP84YAFzXQbFI5pq4UWJy8X8UrV6zzlLeL5DBI9SQaSWP8Qs06p4bKwBYzd2VprkTkoGzRIq
Gyi6L8m5r49Z8K6xMLF3UNwjP2HdElfYlPf+fYZ/AgAA///8V1dhXQUAAA==
`,
	},

	"/": {
		isDir: true,
		local: "static",
	},
}
