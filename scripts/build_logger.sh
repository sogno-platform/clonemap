#!/bin/bash

cd ..
echo "Build Logger Docker Image:" && echo "" \
&& docker build -f build/docker/logger/Dockerfile -t logger . \
