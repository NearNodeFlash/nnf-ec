#!/bin/bash

if ! screen -list | grep -q "pax0"; then
   screen -dmS pax0 /dev/ttyS9 230400
fi

i=0
while [ $i -le 2 ]
do
   if [ ! -f pax0.log ]; then
      screen -S pax0 -X colon "logfile pax0.log^M" &&
      screen -S pax0 -X colon "logfile flush 1^M" &&
      screen -S pax0 -X colon "log on^M"
      sleep 0.3s
   fi
   ((i++))
done

if ! screen -list | grep -q "pax1"; then
   screen -dmS pax1 /dev/ttyS11 230400
fi

i=0
while [ $i -le 2 ]
do
   if [ ! -f pax1.log ]; then
      screen -S pax1 -X colon "logfile pax1.log^M" &&
      screen -S pax1 -X colon "logfile flush 1^M" &&
      screen -S pax1 -X colon "log on^M"
      sleep 0.3s
   fi
   ((i++))
done

screen -S pax0 -X colon "wrap off^M" &&
screen -S pax0 -X stuff "lnkstat\\n" &&
sleep 1s &&
screen -S pax0 -X hardcopy &&
cat hardcopy.0 && rm hardcopy.0
sleep 0.3s

screen -S pax1 -X colon "wrap off^M" &&
screen -S pax1 -X stuff "lnkstat\\n" &&
sleep 1s &&
screen -S pax1 -X hardcopy &&
cat hardcopy.0 && rm hardcopy.0

