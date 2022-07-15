#!/bin/bash

if [ $LOG_LEVEL == 'TRACE' ] || [ $LOG_LEVEL == 'trace' ]; then
    LOG_LEVEL=5
elif [ $LOG_LEVEL == 'DEBUG' ] || [ $LOG_LEVEL == 'debug' ]; then
    LOG_LEVEL=4
elif [ $LOG_LEVEL == 'INFO' ] || [ $LOG_LEVEL == 'info' ]; then
    LOG_LEVEL=3
elif [ $LOG_LEVEL == 'WARN' ] || [ $LOG_LEVEL == 'warn' ]; then
    LOG_LEVEL=2
elif [ $LOG_LEVEL == 'ERROR' ] || [ $LOG_LEVEL == 'error' ]; then
    LOG_LEVEL=1
elif [ $LOG_LEVEL == 'FATAL' ] || [ $LOG_LEVEL == 'fatal' ]; then
    LOG_LEVEL=0
else
    LOG_LEVEL=3
fi

./main -sidecarImage=$SIDECAR_IMAGE -v=$LOG_LEVEL