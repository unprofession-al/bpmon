---
title: "Persist Historical Data"
date: 2018-10-29T11:07:36+01:00
optional: true
duration: "10 min"
menu: "doc"
weight: 40
---

See how you can track the history of your incidents.

<!--more-->

## Write data to your Influx Database

Since sometimes we need to keep track of our uptimes and incidents in order to report to customers etc. For that
matter BPMON provides the `write`. Given that the `default.store.connection` string is properly set, we can simply
write the data into the database by running:

```
bpmon write
```

BPMON does not come with a daemon mode or similar. In order to have a decent history of your business processes
simply run `bpmon write` with a scheduler such as [cron job](http://man7.org/linux/man-pages/man8/cron.8.html), 
[systemd.timer](https://www.freedesktop.org/software/systemd/man/systemd.timer.html), [Jenkins](https://jenkins.io/), 
via [GitLab](https://about.gitlab.com/) or whatever you have at your disposal.

## Explore your data

Since Influx provides very simple interfaces to access your data such as its HTTP API you have a ton of possibilities
to explore your data. Two very simple approches are the `dashboard` subcommand as well a [Grafana](https://grafana.com/)
dashboard

### The `dashboard` Subcommand

_(coming soon)_

### The Grafana Dashboard

A easy and comfortable to explore BPMON data (or in fact any data stored in an Influx database et al.) is Grafana. You'll
find a example dashboard in our [GitHub repositoy](https://github.com/unprofession-al/bpmon/blob/master/hacking/grafana/BusinessProcessesDashboard.json)

![Grafana Dashboard](images/grafana_dashboard.png "Grafana Dashboard")
