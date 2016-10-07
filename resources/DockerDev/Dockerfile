# Dockerfile
#
# Linux development environment
#
# @author      Nicola Asuni <nicola.asuni@miracl.com>
# ------------------------------------------------------------------------------

FROM phusion/baseimage
MAINTAINER nicola.asuni@miracl.com

ENV DEBIAN_FRONTEND noninteractive
ENV TERM linux
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
ENV HOME /root
ENV DISPLAY :0
ENV GOPATH=/root
ENV PATH=/usr/local/go/bin:$GOPATH/bin:$PATH

# add repositories and update
RUN apt-get update && \
apt-get -y dist-upgrade && \
apt-get install -y apt-utils software-properties-common python-software-properties && \
apt-add-repository universe && \
apt-add-repository multiverse && \
apt-get update

# install development packages and debugging tools
RUN apt-get install -y \
alien \
astyle \
autoconf \
autotools-dev \
build-essential \
cmake \
curl \
debhelper \
devscripts \
dh-golang \
dh-make \
doxygen \
doxygen-latex \
flawfinder \
g++ \
gcc \
gdb \
git \
golang \
golang-golang-x-tools \
lcov \
libc6 \
libc6-dev \
libffi-dev \
lsof \
make \
nano \
parallel \
pbuilder \
pkg-config \
python \
python-dev \
python-pip \
rpm \
rsync \
screen \
strace \
swig \
upx-ucl \
valgrind \
vim \
wget

# Configure SSH
RUN mkdir -p /root/.ssh && echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config

# Install and configure GO
RUN wget https://storage.googleapis.com/golang/go1.7.1.linux-amd64.tar.gz && \
tar xvf go1.7.1.linux-amd64.tar.gz && \
mv go /usr/local && \
mkdir -p /root/bin && \
mkdir -p /root/pkg && \
mkdir -p /root/src && \
echo 'export GOPATH=/root' >> /root/.profile && \
echo 'export PATH=/usr/local/go/bin:$GOPATH/bin:$PATH' >> /root/.profile && \
go version
