#!/bin/sh
#
# dockerbuild.sh
#
# Build the software in a Docker container
#
# @author Nicola Asuni <nicola.asuni@miracl.com>
# ------------------------------------------------------------------------------

# NOTES:
#
# This script requires Docker

# EXAMPLE USAGE:
# ./dockerbuild.sh

# build the environment
docker build -t miracl/godev ./resources/DockerDev/

# generate a docker file on the fly
cat > Dockerfile <<- EOM
FROM miracl/godev
MAINTAINER nicola.asuni@miracl.com
RUN mkdir -p /root/.ssh
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config
RUN mkdir -p /root/GO/src/github.com/miracl/natstest
ADD ./ /root/GO/src/github.com/miracl/natstest
WORKDIR /root/GO/src/github.com/miracl/natstest
RUN make deps && make qa && make build
EOM

# docker image name
DOCKER_IMAGE_NAME="local/build:natstest"

# build the docker container and build the project
docker build --no-cache -t ${DOCKER_IMAGE_NAME} .

# start a container using the newly created docker image
CONTAINER_ID=$(docker run -d ${DOCKER_IMAGE_NAME})

# copy the artifact back to the host
docker cp ${CONTAINER_ID}:"/root/GO/src/github.com/miracl/natstest/target" ./

# remove the container and image
docker rm -f ${CONTAINER_ID} || true
docker rmi -f ${DOCKER_IMAGE_NAME} || true
