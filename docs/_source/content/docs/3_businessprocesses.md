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

## Adding our first Business Process Definition

Now that BPMON is set up, lets define a *business process*. Again, we do that 
via YAML, a file per *business process*. Put the following content into `$BPMON_BASE/bp.d/web_service_x.yaml`

```yaml
---
# Give it a name. Names can be changed anytime...
name: Web Service X
# Also give it an ID. This is used to store results in the database and
# therefore should not be changed.
id: ws_x
# Tell BPMON during what time the process needs to be avalable. Remember
# the availabilities section from the global configuration...? This links 
# there.
availability: medium
# You can also specify a 'responsible' string. This string can then be used in
# the trigger template. This could be for example trigger a specific http
# end point, pass some uri parameters, send an email to a specific address etc.
# The 'responsible' string is inherited by its KPIs if not overwritten...
responsible: app.team@example.com
# By providing a list of 'recipients' subcommands such as 'dashboard' can
# use that information in order do provide some sort of authorization.
recipients: [ UsersAppX ]
# Now the KPIs...
kpis:
  - 
    # We already know the name and ID part...
    name: Database Availability
    id: db_availability
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
      - { host: database1.example.com, service: ping } 
      - { host: database2.example.com, service: ping }
  - name: Frontend Nodes Availability
    id: frontend_availability
    operation: MINPERCENT 60
    services:
      - host: frontend1.example.com
        service: api_health
        responsible: engineering.team@example.com
      - { host: frontend2.example.com, service: api_health }
      - { host: frontend3.example.com, service: api_health }
      - { host: frontend4.example.com, service: api_health }
      - { host: frontend5.example.com, service: api_health }
      - { host: frontend6.example.com, service: api_health }
```

Certainly you have to adopt the configuration to match systems monitored via your icinga instance or use
[icingamock](//github.com/unprofession-al/bpmon/blob/master/cmd/icingamock/README.md) to use our Business Process
Definition:

```
icingamock -bp $BPMON_BASE/bp.d
```

Configuration done, lets check...!
