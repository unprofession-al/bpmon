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

{{% info headline="Keep your system tidy" %}}
Please make sure you place all your config files in a dedicated directory. We will refer to this 
configuration folder as `$BPMON_BASE`. 

Note that you can export the `BPMON_BASE` environment variable or use the `-b/--base` flag to point 
`BPMON` to your own configuration directory (by default `BPMON` expects to find its configuration in
the current directory).
{{% /info %}}

## Prepare the directory structure

In `$BPMON_BASE` run:

```
mkdir bp.d && mkdir runners
```

## Generate the Main Configuration

BPMON provides a couple of sub commands that allow you to manage your main configuration file.
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

Pipe this output in a file called `config.yaml`. 

```
bpmon config init > $BPMON_BASE/config.yaml
```

## Connect to Icinga and Influx database

To get some insights on what can be configured please read the comment in 
this generated file. For now we only need to setup the `checker` and `store` parts of the configuration to get things started.

In `default.checker.connection` add the connection string to access your icinga API...

In `default.store` we have two options:

1. If you have an Influx database ready paste the connection string at `default.store.connection`.
2. If you don't want to persist historical data right now set `default.store.get_last_status` to false. Add `http://in.existent` 
   at `default.store.connection`.

## Define an availability

Often we have some time slots in which the availability of a system is guaranteed. Add those time slots to your main configuration in `default.availabilities`:

```
---
default:
  ...
  availabilities:
    high:
      monday:    [ "allday" ]
      tuesday:   [ "allday" ]
      wednesday: [ "allday" ]
      thursday:  [ "allday" ]
      friday:    [ "allday" ]
      saturday:  [ "allday" ]
      sunday:    [ "allday" ]
    medium:
      monday:    [ "06:00:00-20:00:00" ]
      tuesday:   [ "06:00:00-20:00:00" ]
      wednesday: [ "06:00:00-20:00:00" ]
      thursday:  [ "06:00:00-20:00:00" ]
      friday:    [ "06:00:00-20:00:00" ]
      saturday:  [ "06:00:00-20:00:00" ]
      sunday:    [ "06:00:00-20:00:00" ]
    low:
      monday:    [ "08:00:00-12:00:00", "13:30:00-17:00:00" ]
      tuesday:   [ "08:00:00-12:00:00", "13:30:00-17:00:00" ]
      wednesday: [ "08:00:00-12:00:00", "13:30:00-17:00:00" ]
      thursday:  [ "08:00:00-12:00:00", "13:30:00-17:00:00" ]
      friday:    [ "08:00:00-12:00:00", "13:30:00-17:00:00" ]
  ...
```

In this case we have three availabilities defined: 'high', 'medium', 'low'. Name yours however your want, just make sure the names make sense to you.

That's it for the main configuration! Let's move on...
