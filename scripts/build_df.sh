#!/bin/bash

cd ..
echo "Build DF Docker Image:" && echo "" \
&& docker build -f build/docker/df/Dockerfile -t df . \
