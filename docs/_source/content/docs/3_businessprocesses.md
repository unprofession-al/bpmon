---
title: "Create Business Processes"
date: 2018-10-29T11:07:36+01:00
optional: false
duration: "10 min"
menu: "doc"
weight: 20
---

Learn how do create a Business Process composed from existing Icinga2 Checks.

<!--more-->

## Adding our first Business Process

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
