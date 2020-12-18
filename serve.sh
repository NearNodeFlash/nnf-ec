#!/bin/bash

kubectl port-forward $(kubectl get pods | grep nnf-ec | awk '{ print $1 }') 8080:8080 