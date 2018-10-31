---
title: "Run BPMON"
date: 2018-10-29T11:07:36+01:00
optional: false
duration: "2 min"
menu: "doc"
weight: 30
---

All set and configured... Let's see the results!

<!--more-->

## Here we go...

In your `$BPMON_BASE` directory simply run: 

``` bash
$ bpmon run    
  Web Service X is ok
    Database Availability is ok
      database1.example.com!ping is ok
      database2.example.com!ping is ok
    Frontend Nodes Availability is ok
      frontend3.example.com!api_health is ok
      frontend5.example.com!api_health is ok
      frontend4.example.com!api_health is ok
      frontend2.example.com!api_health is ok
      frontend1.example.com!api_health is ok
      frontend6.example.com!api_health is ok

```

You just verified that your first business process is up and running!
