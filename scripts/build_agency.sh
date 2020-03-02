#!/bin/bash

cd ..
echo "Build Agency Docker Image:" && echo "" \
&& docker build -f build/docker/agency/Dockerfile -t agency . \
