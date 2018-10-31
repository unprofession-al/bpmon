---
title: "Write your own Runners"
date: 2018-10-29T11:07:36+01:00
optional: true
duration: "15 min"
menu: "doc"
weight: 50
---

Not quite what you need yet? Let's customize the output!

<!--more-->

## Runners?

Whenever you type `bpmon run` you execute a _Runner_. This is nothing but a [go text/template](https://golang.org/pkg/text/template/) 
and some meta data. There are default runner (such as the one named `default` which will run if no runner is specified) which you can list 
via the `--list` flag:

```
$ bpmon run --list
verbose
	Print all check results in a long format and human readable
issues
	Print failed check results in a short format and human readable
issues_verbose
	Print failed check results in a long format and human readable
default
	Print all check results in a short format and human readable
```

Try them out by specifying the Runner you want to execute. For example `bpmon run issues` will only print failed business processes, 
`bpmon run verbose` prints a ton of details.

If you want to dig into the details of these predefined runners have a look at their
[source](https://github.com/unprofession-al/bpmon/blob/master/runners/defaults.go).

Since we are obviousely only rendering some templates we can easily build our own runners... 

## Learning by Example: Generate a status.html with a Runner

Wouldn't it be nice to render some web page that shows the current business process status? Here we go...

Create two files in the directory `$BPMON_BASE/runners/status/`. The first file `cmd.yaml` contains some meta data:

``` yaml
---
description: |
  This renders a html file to stdout that represents the current status.
```

The second file `cmd.template` it the template:

``` html
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>status.html</title>
        <style>
.content { width: 70%; margin: 0 auto; }
.bp { margin: 20px; padding: 20px; padding-right: 0; }
.kpi { padding: 10px; padding-left: 30px; padding-right: 0; padding-bottom: 20px; margin-top: 10px; }
.svc { padding: 10px; padding-left: 30px; padding-right: 0; margin-top: 10px; }
.bp.ok { background-color: #8cd98c; }
.kpi.ok { background-color: #66cc66; }
.svc.ok { background-color: #40bf40; }
.bp.nok { background-color: #ff6666; }
.kpi.nok { background-color: #ff3333; }
.svc.nok { background-color: #ff0000; }
.bp.unknown { background-color: #d966ff; }
.kpi.unknown { background-color: #cc33ff; }
.svc.unknown { background-color: #bf00ff; }
.bp>.title, .kpi>.title { font-weight: 700; }
.kpi>.title, .svc>.title { font-size: 14px; }
        </style>
    </head>
    <body>
        <div class="content">
            {{ range $index, $bp := .BP }}
            {{$status := "unknown"}}{{if eq $bp.Status 0}}{{$status = "ok"}}{{else if eq $bp.Status 1}}{{$status = "nok"}}{{end}}
            <div class="bp {{$status}}" id="{{$bp.ID}}">
                <div class="title">BP {{$bp.Name}}</div>
                {{- range $index, $kpi := .Children }}
                {{$status := "unknown"}}{{if eq $kpi.Status 0}}{{$status = "ok"}}{{else if eq $kpi.Status 1}}{{$status = "nok"}}{{end}}
                <div class="kpi {{$status}}" id="{{$kpi.ID}}">
                    <div class="title">KPI {{$kpi.Name}}</div>
                    {{- range $index, $svc := .Children }}
                    {{$status := "unknown"}}{{if eq $svc.Status 0}}{{$status = "ok"}}{{else if eq $svc.Status 1}}{{$status = "nok"}}{{end}}
                    <div class="svc {{$status}}" id="{{$svc.ID}}">
                        <div class="title">SVC {{$svc.Name}}</div>
                        {{if ne $status "ok"}}<div class="output">{{$svc.Output}}</div>{{end}}
                    </div>
                    {{- end -}}
                </div>
                {{ end -}}
            </div>
            {{ end -}}
        </div>
    </body>
</html>
```

Wanna see?

```
$ bpmon run status > status.html && firefox status.html; rm status.html
```

Now you can imagine the possibilities. Do you want to trigger pagerduty via API if a business process is failed? Create a runner that renders a bash 
script and schedule via `bpmon run myrunner | bash`. 

## Meet your friends

Now the biggest hassle when buildung templates is to know the data you have available. To give you a hand we have added some funcions to
the templating engine to help you inspect the data:

| Function | Description | Usage |
| --- | --- | --- |
| pretty | Print the data in a human readable format | `{{ pretty . }}` |
| json | Print the data as JSON object | `{{ json . }}` |
| yaml | Print as YAML object | `{{ yaml . }}` |
| spew | Print with a lot of details regarding data type and such | `{{ spew . }}` |
| describe | Print the data structure rather than the data itself | `{{ describe . }}` |

Remember those functions, they might get handy.
