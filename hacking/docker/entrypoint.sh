#!/bin/sh

ARGS="-c /bpmon_config/conf.yaml -b /bpmon_config/bp.d/"

if [[ -z "${CMD}" ]]; then
    ARGS="${ARGS} run"
else
    ARGS="${ARGS} ${CMD}"
fi

if [[ ! -z "${STATIC}" ]]; then
    ARGS="${ARGS} --static ${STATIC}"
fi

if [[ ! -z "${PEPPER}" ]]; then
    ARGS="${ARGS} --pepper ${PEPPER}"
fi

./bpmon ${ARGS}
