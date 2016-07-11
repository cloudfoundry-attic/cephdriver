#!/bin/bash
"${DRIVERS_PATH:?DRIVERS_PATH must be set}"

cd `dirname $0`

pkill -f cephdriver

mkdir -p ../mountdir

if [ $TRANSPORT == "tcp" ];
then
echo "RUNNING TCP ACCEPTANCE"
../exec/cephdriver -listenAddr="0.0.0.0:9750" -transport="tcp" -driversPath="$DRIVERS_PATH" &
else
echo "RUNNING UNIX ACCEPTANCE"
../exec/cephdriver -listenAddr="${driversPath}/cephdriver.sock" -transport="unix" -driversPath="$DRIVERS_PATH" &
fi