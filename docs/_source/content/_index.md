---
title: "Home"
date: 2018-10-29T11:07:36+01:00
menu: "main"
---

Business Process Monitor that uses Icinga as service health source. BPMON allows
you to compose *Business Processes* (BPs) and their *Key Performance Indicators*
(KPIs) using the services you already have in Icinga! BPMON ...

* ... *reads* the states of services defined, ...
* ... *evaluates* those states, ...
* ... feeds the results results to an InfluxDB in order to *keep track of the history* ...
* ... and/or triggers a command to perform actions such as *alarming*.

BPMON does not need to be run as a server. Run BPMON via Jenkins, Cron or
manually as needed.

## Installation

To install BPMON, you need Go 1.7. Please refer to https://golang.org/doc/install 
to do so...

Simply run the following command (don't forget the three dots at the end!):
```
# go get -u github.com/unprofession-al/bpmon/...
```

## Configure BPMON

BPMONs global settings are configured via a simple config file written in YAML.
The config file can hold multiple *sections* (separat configurations) that enables 
a decent amount of flexibility in combination with YAMLs *anchors*. Let's have 
a look at an example:

```yaml
---
# The default section is - as suggested - read by default. Note the '&default'
# notation; this is known an an 'anchor' and allows you to reuse the settings
# in other sections...
default: &default
  # First BPMON needs to have access to your Icinga2 API. Learn more on by reading 
  # https://docs.icinga.com/icinga2/latest/doc/module/icinga2/chapter/icinga2-api.
  checker:
    kind: icinga
    connection: https://user:pass@icinga.example.com:5665
  # Also the connection to the InfluxDB is required in order to persist the
  # state for reporting and such. 
  store:
    kind: influx
    connection: http://user:pass@influx.example.com:8086/bpmon
    # the timeout is read as a go (golang) duration, please refer to 
    # https://golang.org/pkg/time/#Duration for a detailed explanation.
    timeout: 2s
    # If a state is 'OK' only save it to InfluxDB if its an BP measurement 
    # (e.g. do not persist 'OK' states for KPIs and services for the sake of
    # a small amount of data). In that case BP 'OK' states are saved as 
    # 'heart beat' of BPMON.
    save_ok: [ BP ]
    # This will tell BPMON to compare the current status against the last 
    # status saved in InfluxDB and adds some values to the measurement 
    # accordingly. This then allows to generate reports such as 'Tell me
    # only when a status is changed from good to bad'. This only runs against
    # types listed in 'save_ok' since only these are persisted 'correctly'.
    get_last_status: true
  # global_recipient will be added to the repicient list af all BP
  global_recipient: admin
  # Define your office hours et al. according to your service level 
  # agreements (SLA). You can later reference them in your BP definitions.
  availabilities:
    7x24:
      monday:    [ "allday" ]
      tuesday:   [ "allday" ]
      wednesday: [ "allday" ]
      thursday:  [ "allday" ]
      friday:    [ "allday" ]
      saturday:  [ "allday" ]
      sunday:    [ "allday" ]
    9to5:
      monday:    [ "09:00:00-17:00:00" ]
      tuesday:   [ "09:00:00-17:00:00" ]
      wednesday: [ "09:00:00-17:00:00" ]
      thursday:  [ "09:00:00-17:00:00" ]
      friday:    [ "09:00:00-17:00:00" ]
      saturday:  [ "09:00:00-17:00:00" ]
      sunday:    [ "09:00:00-17:00:00" ]
test:
  # Inherit all settings fron the default anchor and extend/overwrite
  <<: *default
  availabilities:
    officeHours:
      monday:    [ "09:00:00-12:00:00", "13:30:00-17:00:00" ]
      tuesday:   [ "09:00:00-12:00:00", "13:30:00-17:00:00" ]
      wednesday: [ "09:00:00-12:00:00", "13:30:00-17:00:00" ]
      thursday:  [ "09:00:00-12:00:00", "13:30:00-17:00:00" ]
      friday:    [ "09:00:00-12:00:00", "13:30:00-17:00:00" ]
alarming:
  <<: *default
  # Extend the default rules; in that case: Do not run the alarming command
  # if a critical service is aready aknowledged to avoid alarm spamming.
  rules:
    25:
      must: [ critical, acknowledged ]
      then: ok
  influx:
    get_last_status: false
  # If a service is failed, this command (rendered as a golang template) is 
  # printed to the stdout. This allows to easily wrap BPMON into an eval
  # statement in your shell script.
  trigger:
    template: |
        curl -X POST https://alarming.example.com/trigger -d '[{{- range $index, $elem := . -}}
          {{- if $index }},{{ end -}}
            {{- "{" }} "name": "{{ .Name }}", "services": ["
              {{- range $index, $elem := .Children -}}
                {{- if $index}},"{{ end -}}
                {{- range $index, $elem := .Children -}}
                  {{- if $index }}","{{ end -}}
                  {{- $elem.Name -}}
                {{- end -}}"
              {{- end -}}]
            {{- "}" -}}
          {{- end -}}]'
```

Run the following command to check your configuration and see how it comes
together with the defaults set by BPMON:

```
# bpmon -c [path_to_the_config] -s [name_of_the_section_to_check] config
```

For example:

```
# bpmon -c ./cfg.yaml -s default config
```

## Define *Business Processes*

Now that BPMON is set up, lets define a *business process*. Again, we do that 
via YAML, a file per *business process*:

```yaml
---
# Give it a name. Names can be changed anytime...
name: Application X
# Also give it an ID. This is used to store results in the database and
# therefore should not be changed.
id: app_x
# Tell BPMON during what time the process needs to be avalable. Remember
# the availabilities section from the global configuration...? This links 
# there.
availability: 9to5
# You can also specify a 'responsible' string. This string can then be used in
# the trigger template. This could be for example trigger a specific http
# end point, pass some uri parameters, send an email to a specific address etc.
# The 'responsible' string is inherited by its KPIs if not overwritten...
responsible: app.team@example.com
# By providing a list of 'recipients' subcommands such as 'beta dashboard' can
# use that information in order do provide some sort of authorization.
recipients: [ UsersAppX ]
# Now the KPIs...
kpis:
  - 
    # We already know the name and ID part...
    name: Load Balancer Availability
    id: lb_availability
    # The 'operatinon' defines how the services must be evaluated. Possible
    # options are:
    # * AND:          All services need to be 'OK' for the KPI to be 'OK'.
    # * OR:           At least one sf its services needs to fo 'OK'.
    # * MIN x:        Where x is an integer. A minimum number of x services
    #                 need to be 'OK'
    # * MINPERCENT x:  As 'MIN', but in percent.
    operation: OR
    # Again, a 'responsible' string can be specified in order not to inherit
    # from the parent BP.
    responsible: infra.team@example.com
    # And now the processes. Host and service relate to how you named those
    # things in your Icinga2 setup.
    services:
      - { host: haproxy1.example.com, service: ping } 
      - { host: haproxy2.example.com, service: ping }
  - name: App Nodes Availability
    id: app_availability
    operation: MINPERCENT 50
    services:
      - { host: app1.example.com, service: api_health, responsible: engineering.team@example.com }
      - { host: app2.example.com, service: api_health }
      - { host: app3.example.com, service: api_health }
      - { host: app4.example.com, service: api_health }
```

Place your business processes in a directory and point BPMON there to execute:

```
# bpmon -c ./cfg.yaml -s default -b ./bp.d/ run
```

## Run

```
# bpmon -h
Montior business processes composed of Icinga checks

Usage:
  bpmon [command]

Available Commands:
  beta        Access beta features of BPMON
  config      Print the configurantion used to stdout
  run         Run all business process checks and print to stdout
  trigger     Run all business process checks and trigger a custom command if issue occure
  write       Insert data into InfluxDB

Flags:
  -b, --bp string        path to business process configuration files (default "/etc/bpmon/bp.d")
  -c, --cfg string       path to the configuration file (default "/etc/bpmon/cfg.yaml")
  -p, --pattern string   pattern of business process configuration files to process (default "*.yaml")
  -s, --section string   name of the section to be read (default "default")
  -v, --verbose          print log output (default true)

Use "bpmon [command] --help" for more information about a command.
```
