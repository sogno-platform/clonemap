#!/bin/bash

cd ..
echo "Build AMS Docker Image:" && echo "" \
&& docker build -f build/docker/ams/Dockerfile -t ams . \
