#!/bin/bash
docker stop rfsf-evt
docker rm rfsf-evt 
docker build $1 -t rfsf-event .
docker run -v /var/run/dbus:/var/run/dbus -d --name rfsf-evt -p 5000:8080 rfsf-event
#docker run --privileged -v /var/run/dbus:/var/run/dbus -d --name rfsf-evt -p 5000:8080 rfsf-event
docker attach rfsf-evt
