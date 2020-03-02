#!/bin/bash

cd ..
echo "Build AMS Docker Image:" && echo "" \
&& docker build -f build/docker/ams/Dockerfile -t ams . \
&& echo "" && echo "Tag Image: " && echo "" \
&& docker tag ams registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/ams \
&& echo "" && echo "Push Image to Registry:" && echo "" \
&& docker push registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/ams

if [ $? -eq 0 ]
then
    echo ""
    echo "Image successfully pushed to registry"
else
    echo ""
    echo "Failure"
fi
