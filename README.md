![BPMON](https://raw.githubusercontent.com/unprofession-al/bpmon/master/bpmon.png "BPMON")

Business Process Monitor that uses Icinga as process health source. BPMON allows
you to compose *Business Processes* (BPs) and their *Key Performance Indicators*
(KPIs) using the services you already have in Icinga! BPMON ...

* ... *reads* the states of services defined, ...
* ... *evaluates* those states, ...
* ... feeds the results results to an InfluxDB in order to *keep track of the history* ...
* ... and/or triggers a command to perform actions such as *alarming*.

BPMON does not need to be run as a server. Run BPMON via Jenkins, Cron or
manually as needed.

## Install and Build
    # go get -u github.com/unprofession-al/bpmon/...

## Configure BPMOM

BPMONs global settings are configured via a simple config file written in YAML.
The config file can hold multiple *sections* (separat configurations) that enables 
a decent amount of flexibility in combination with YAMLs *anchors*. Let's have 
a look at an example:

    ---
    # The default section is - as suggested - read by default. Note the '&default'
    # notation; this is known an an 'anchor' and allows you to reuse the settings
    # in other sections...
    default: &default
      # First BPMON needs to have access to your Icinga2 API. Learn more on by reading 
      # https://docs.icinga.com/icinga2/latest/doc/module/icinga2/chapter/icinga2-api.
      icinga:
        server: icinga.example.com
        port: 5665
        pass: youllneverguess
        user: bpmon
        proto: https
      # Also the connection to the InfluxDB is required in order to persist the
      # state for reporting and such. 
      influx:
        connection:
          server: ***REMOVED***
          port: 8086
          proto: http
        database: bpmon
        # If a state is 'OK' only save it to InfluxDB if its an BP measurement 
        # (e.g. do not persist 'OK' states for KPIs and services for the sake of
        # a small amount of data). In that case BP 'OK' states are saved as 
        # 'heart beat' of BPMON.
        save_ok: [ BP ]
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
      influx:
        database: bpmont
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
      # if a critical service is aready aknowledged to avoin alarm spamming.
      rules:
        25:
          must: [ critical, acknowledged ]
          then: ok
      # If a Service is failed, this command (rendered as a golang template) is 
      # printed to the stdout. This allows to easily wrap BPMON into an eval
      # statement in your shell script.
      trigger:
        template: |
            curl -X POST -u ***REMOVED*** https://***REMOVED***/alerts/new\?trigger_alert\=1 -d '[{{- range $index, $elem := . -}}
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

## Define *Business Processes*

## Run
    # bpmon -h
    Montior business processes composed of Icinga checks

    Usage:
      bpmon [command]

    Available Commands:
      config      Print the configurantion used to stdout
      run         Run all business process checks and print to stdout
      trigger     Run all business process checks and trigger a custom command if issue occure
      write       Insert data into InfluxDB

    Flags:
      -b, --bp string        path to business process configuration files (default "/etc/bpmon/bp.d")
      -c, --cfg string       path to the configuration file (default "/etc/bpmon/cfg.yaml")
      -p, --pattern string   pattern of business process configuration files to process (default "*.yaml")
      -s, --section string   name of the section to be read (default "default")

    Use "bpmon [command] --help" for more information about a command.

