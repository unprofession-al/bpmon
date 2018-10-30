---
title: "Home"
date: 2018-10-29T11:07:36+01:00
menu: "main"
weight: 1
---

## 'BPMON' what?

> BPMON is a tool that lets you monitor Business Processes composed from the checks
in your monitoring system.

## Use Case

Often in IT a couple of tiny indicators we have in our monitoring system state
weither an important service we provide is running or not. Let


Business Process Monitor that uses Icinga as service health source. BPMON allows
you to compose *Business Processes* (BPs) and their *Key Performance Indicators*
(KPIs) using the services you already have in Icinga! BPMON ...

* ... *reads* the states of services defined, ...
* ... *evaluates* those states, ...
* ... feeds the results results to an InfluxDB in order to *keep track of the history* ...
* ... and/or triggers a command to perform actions such as *alarming*.

BPMON does not need to be run as a server. Run BPMON via Jenkins, Cron or
manually as needed.


