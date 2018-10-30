---
title: "The Main Configuration"
date: 2018-10-29T11:07:36+01:00
menu: "doc"
optional: false
duration: "5 min"
weight: 10
---

Now let's move forward; here comes the fun part.

<!--more-->

### Generate the Main Configuration

BPMON provides a couple of subcommands that allow you to manage your main configuration file.
When starting a new setup from scratch `bpmon config init` comes handy. This will print an annotated 
configuration file prefilled with defaults to the standard output:

```
$ bpmon config init
# The default section is - as the name suggests - read by default. Note the '&default'
# notation; this is known an an 'anchor' and allows you to reuse the settings
# in other sections...
default:
  # global_recipients will be added to the repicients list af all BP
  global_recipients: []
  # health ... TODO
  health:
    store_required: false
    checker_required: true
    responsible: ""
    name: ""
    id: bla
  # First BPMON needs to have access to your Icinga2 API. Learn more on by reading
  # https://docs.icinga.com/icinga2/latest/doc/module/icinga2/chapter/icinga2-api.
  checker:
    # kind defines the checker implementation to be used by BPMON. Currently
    # only icinga is implemented.
    kind: icinga
    # The connection string describes how to connect to your Icinga API. The
    # string needs to follow the pattern:
    #   [protocol]://[user]:[passwd]@[hostname]:[port]
    connection: ""
    # BPMON verifies if a https connection is trusted. If you wont to trust a
    # connection with an invalid certificate you have to set this to true.
    tls_skip_verify: false
    # timeout defines how long BPMON waits for each request to the checker to
    # recieve a response. The string is parsed as a goland duration, refer to
    # its documentation for more details:
    #   https://golang.org/pkg/time/#ParseDuration
    timeout: 10s
  # Also the connection to the InfluxDB is required in order to persist the
  # state for reporting and such
  store:
    kind: influx
    connection: ""
    timeout: 10s
    save_ok:
    - BP
    get_last_status: true
    debug: false
    # BPMON verifies if a https connection is trusted. If you wont to trust a
    # connection with an invalid certificate you have to set this to true
    tls_skip_verify: false
  # Define your office hours et al. according to your service level
  # agreements (SLA). You can later reference them in your BP definitions.
  availabilities: {}
  # Extend the default rules; in that case: Do not run the alarming command
  # if a critical service is aready aknowledged to avoid alarm spamming.
  rules: {}
  # dashboard configures the dashboard subcommand.
  dashboard:
    # listener tells the dashboard where to bind. This string
    # should match the pattern [ip]:[port].
    listener: 127.0.0.1:8910
    # static is the path to the directory that sould be served
    # at the root of the server. This should contain the UI of the
    # Dashboard
    static: ""
    # grant_write is a list of recipients which are allowed to access the annotate
    # endpoint via POST request.
    grant_write: []
  env:
    runner: runners/
    bp: bp.d/
```
