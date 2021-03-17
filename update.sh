#!/bin/bash

cd deployments/docker
docker-compose down
docker rmi qianwen12/frontend:latest
docker rmi frontend:latest
cd ../..
docker build -f build/docker/frontend/Dockerfile -t frontend .
docker tag frontend:latest qianwen12/frontend:latest
docker push qianwen12/frontend:latest
cd deployments/docker
docker-compose up


