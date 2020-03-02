#!/bin/bash

cd ..
echo "Build Logger Docker Image:" && echo "" \
&& docker build -f build/docker/df/Dockerfile -t df . \
&& echo "" && echo "Tag Image: " && echo "" \
&& docker tag df registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/df \
&& echo "" && echo "Push Image to Registry:" && echo "" \
&& docker push registry.git.rwth-aachen.de/acs/research/ensure/clonemap/clonemap/df

if [ $? -eq 0 ]
then
    echo ""
    echo "Image successfully pushed to registry"
else
    echo ""
    echo "Failure"
fi
