#!/bin/bash
x=1
y=false
while [ $x -le 90 ]; do
  if [ "$(echo PING | nc -U /var/run/clamav/clamd.ctl)" == "PONG" ]; then exit 0; fi
  sleep 5
  x=$(($x + 1))
done
exit 1
