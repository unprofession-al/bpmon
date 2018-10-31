---
title: "Getting Started with BPMON"
date: 2018-10-29T11:07:36+01:00
optional: false
duration: "5 min"
menu: "doc"
weight: 1
---

There are a few things you need to prepare... Getting everything ready.

<!--more-->

## Installation

BPMON itself consists of a single binary with no dependencies. There are a couple of options to get a BPMON binary
on your system:

### From Source

To install BPMON, you need Go 1.11.x. Please refer to [the official documentation](https://golang.org/doc/install) 
to do so...

As soon as your Go environment is setup simply run the following command (don't forget the three dots at the end!):

```
# go get -u github.com/unprofession-al/bpmon/...
```

This will fetch the source and its compile time dependencies and install it under `$GOPATH/bin/bpmon`

### Get a Binary Release

*(coming soon)*

### Via Docker Hub

*(coming soon)*

## Preparing ICINGA 2 API Access

<div class="info">
    <div class="headline">No ICINCA 2 Server available? No time or permission to setup the API?</div>
    <p>
        No problem. BPMON comes with a tiny <a href="https://github.com/unprofession-al/bpmon/tree/master/cmd/icingamock/README.md">Icinga Mock Server</a> to get your hands dirty without having ICINGA
        ready... 
    </p>
</div>

BPMON fetches the status of the required services via the ICINGA 2 API. Therefore we have to enable the API as well as 
create a user for BPMON. Refer to the [official documentation](https://icinga.com/docs/icinga2/latest/) to do so... 

1. [Setting up the API](https://icinga.com/docs/icinga2/latest/doc/12-icinga2-api/#setting-up-the-api)
2. [Creating an ApiUser](https://icinga.com/docs/icinga2/latest/doc/12-icinga2-api/#authentication)

Make sure you apply the correct permissions:

```
object ApiUser "bpmon" {
  password = "..."
  permissions = ["objects/query/Host","objects/query/Service","status/*"]
}
```

## Setting up an Influx Database (optional)

A feature of bpmon is to write all measurements in an Influx database on order to have historical data of our up- and
downtimes as well as the reasons for potential incidents. This is a neat feature for reporting etc.

If you want to get your hands on this feature you need to have an Influx database as well as a username/password 
with read/write access at hand.

Visit their [documentaiton](https://docs.influxdata.com/influxdb/) to learn how to set things up.

