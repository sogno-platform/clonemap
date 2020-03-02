#!/bin/bash

cd ..
echo "Build agency Docker Image:" && echo "" \
&& docker build -f build/docker/agency/Dockerfile -t agency . \
&& echo "" && echo "Tag Image: " && echo "" \
&& docker tag registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/agency \
&& echo "" && echo "Push Image to Registry:" && echo "" \
&& docker push registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/agency

if [ $? -eq 0 ]
then
    echo ""
    echo "Image successfully pushed to registry"
else
    echo ""
    echo "Failure"
fi
