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

BPMON itself consists of a single binary with no dependencies. There are a couple of options to get a BPMON binary on your system:

### From Source

To install BPMON, you need Go 1.11.x. Please refer to [the official documentation](https://golang.org/doc/install) 
to do so...

As soon as your Go environment is setup simply run the following command (don't forget the three dots at the end!):

```
# go get -u github.com/unprofession-al/bpmon/...
```

This will fetch the source and its compile time dependencies and install it under `$GOPATH/bin/bpmon`
