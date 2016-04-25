#!/bin/bash

cd `dirname $0`

pkill -f cephdriver

mkdir -p ~/voldriver_plugins

mkdir -p ../mountdir

driversPath=$(realpath ~/voldriver_plugins)
#../exec/cephdriver -listenAddr="0.0.0.0:9750" -transport="tcp" -mountDir="../mountdir" -driversPath="${driversPath}" &



if [$TRANSPORT == "tcp"]
then
../exec/cephdriver -listenAddr="0.0.0.0:9750" -transport="tcp" -driversPath="${driversPath}" &
else
../exec/cephdriver -listenAddr="${driversPath}/cephdriver.sock" -transport="unix" -driversPath="${driversPath}" &
fi