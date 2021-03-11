#!/bin/bash

kubectl describe pod $(kubectl get pods | grep nnf-ec | awk '{ print $1 }') 