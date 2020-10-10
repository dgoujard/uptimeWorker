#!/bin/bash

GOOS=linux go build -o uptimeWorkerLambda github.com/dgoujard/uptimeWorker/cmd/uptimeWorkerLambda && zip function.zip  uptimeWorkerLambda && rm uptimeWorkerLambda