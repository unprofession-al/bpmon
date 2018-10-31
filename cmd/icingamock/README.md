# ICINGAMOCK

`icingamock` is a very simple ICINGA 2 API mock server built specifically to test
and debug [BPMON](//bpmon.unprofession.al). 

It also provides a basic web frontend to toggle the state of your serivces. 

## Install

If you have BPMON installed chances are you also have `icingamock` in your `$GOPATH`. If not, 
just go and get BPMON which will also install `icingamock`:

```
go get -u github.com/unprofession-al/bpmon/...
```

Your are done.

## Configure your 'checks'

There are two ways to get `icingamock` to simulate your environment (eg. have your checks ready): 
You can either define your envionment in a `environment setup file` or your can reference 
your exinting `business process definitions` you already built for BPMON...

### Environment Setup File

Create a directory to place your environment setup files:

``` bash
mkdir ./env.d/ && cd ./env.d/
```

Now create your first environment named `demo` by creating a file named `demo.yaml`:

``` yaml
---
# Our first host is called mysql-01.demo.io
mysql-01.demo.io:
  # The host has two services, one is up (check_state: 0)
  # one is down (check_state: 1) on start
  mysql:
    check_state: 0
    check_output: "MySQL is up and running"
    acknowledgement: false
    downtime: false
  mysql-backup:
    check_state: 1
    check_output: "backup job is failed"
    acknowledgement: true
    downtime: true
```

Add as many hosts/services as you whish to the file

Run `icingamock` and point to the directory your created: 

``` bash
icingamock -env ./env.d/
```

This will spin up the mock server which you can access locally via http://0.0.0.0:8765/?env=demo (note the env query parameter at the end).

Use the following connection string as `checker.connection` in your BPMON config: `http://0.0.0.0:8765/icinga/demo`

### BPMON Business Process Definitions

If you already have some BP definitions for BPMON ready you can refer to them in order do create an environent based on those definitions:

``` bash
icingamock -bp $BPMON_BASE/bp.d
```

This will spin up the mock server which you can access locally via http://0.0.0.0:8765/

Use the following connection string as `checker.connection` in your BPMON config: `http://0.0.0.0:8765/icinga/_`

## Further options

Run `icingamock -h` to see all options:

```
$ icingamock -h 
Usage of icingamock:
  -bp string
    	bpmon bp files
  -env string
    	environment setup files
  -listener string
    	ip/port to listen on (default "0.0.0.0:8765")
  -static string
    	static html served at http root
```
