---
title: "Home"
date: 2018-10-29T11:07:36+01:00
menu: "main"
weight: 1
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

Run the following command to check your configuration and see how it comes
together with the defaults set by BPMON:

```
# bpmon -c [path_to_the_config] -s [name_of_the_section_to_check] config
```

For example:


## Define *Business Processes*

Now that BPMON is set up, lets define a *business process*. Again, we do that 
via YAML, a file per *business process*:
Place your business processes in a directory and point BPMON there to execute:

```
# bpmon -c ./cfg.yaml -s default -b ./bp.d/ run
```

## Run

