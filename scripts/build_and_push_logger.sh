#!/bin/bash

cd ..
echo "Build Logger Docker Image:" && echo "" \
&& docker build -f build/docker/logger/Dockerfile -t logger . \
&& echo "" && echo "Tag Image: " && echo "" \
&& docker tag logger registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/logger \
&& echo "" && echo "Push Image to Registry:" && echo "" \
&& docker push registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/logger

if [ $? -eq 0 ]
then
    echo ""
    echo "Image successfully pushed to registry"
else
    echo ""
    echo "Failure"
fi
