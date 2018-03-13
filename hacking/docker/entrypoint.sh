#!/bin/bash

if [[ ${CMD} == ./bpmon* ]]; then
    eval ${CMD}
else
    ./bpmon ${CMD}
fi
