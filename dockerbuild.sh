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

# project root path
PRJPATH=/root/src/milagro-crypto-c

# generate a docker file on the fly
cat > Dockerfile <<- EOM
FROM miracl/amcldev
MAINTAINER nicola.asuni@miracl.com
RUN mkdir -p ${PRJPATH}
ADD ./ ${PRJPATH}
WORKDIR ${PRJPATH}
RUN make deps && \
make qa && \
make build
EOM

# docker image name
DOCKER_IMAGE_NAME="local/build:natstest"

# build the docker container and build the project
docker build --no-cache -t ${DOCKER_IMAGE_NAME} .

# start a container using the newly created docker image
CONTAINER_ID=$(docker run -d ${DOCKER_IMAGE_NAME})

# copy the artifact back to the host
docker cp ${CONTAINER_ID}:"${PRJPATH}/target" ./

# remove the container and image
docker rm -f ${CONTAINER_ID} || true
docker rmi -f ${DOCKER_IMAGE_NAME} || true
