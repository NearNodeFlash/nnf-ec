#!/bin/bash

kubectl logs $(kubectl get pods | grep nnf-ec | awk '{ print $1 }') -c ${1:-"nnf-ec"}