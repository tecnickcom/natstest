# MAKEFILE
#
# @author      Nicola Asuni <info@tecnick.com>
# @link        https://github.com/tecnickcom/natstest
#
# This file is intended to be executed in a Linux-compatible system.
# It also assumes that the project has been cloned in the right path under GOPATH:
# $GOPATH/src/github.com/tecnickcom/natstest
#
# ------------------------------------------------------------------------------

# Use bash as shell (Note: Ubuntu now uses dash which doesn't support PIPESTATUS).
SHELL=/bin/bash

# CVS path (path to the parent dir containing the project)
CVSPATH=stash.certivox.com/scm/maas

# Project owner
OWNER=MIRACL

# Project vendor
VENDOR=miracl

# Project name
PROJECT=natstest

# Project version
VERSION=$(shell cat VERSION)

# Project release number (packaging build number)
RELEASE=$(shell cat RELEASE)

# Name of RPM or DEB package
PKGNAME=${VENDOR}-${PROJECT}

# Current directory
CURRENTDIR=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))

# GO lang path
ifneq ($(GOPATH),)
	ifeq ($(findstring $(GOPATH),$(CURRENTDIR)),)
		# the defined GOPATH is not valid
		GOPATH=
	endif
endif
ifeq ($(GOPATH),)
	# extract the GOPATH
	GOPATH=$(firstword $(subst /src/, ,$(CURRENTDIR)))
endif

# Add the GO binary dir in the PATH
export PATH := $(GOPATH)/bin:$(PATH)

# Path for binary files (where the executable files will be installed)
BINPATH=usr/bin/

# Path for configuration files
CONFIGPATH=etc/$(PROJECT)/

# Path for init script
INITPATH=etc/init.d/

# Path path for documentation
DOCPATH=usr/share/doc/$(PKGNAME)/

# Path path for man pages
MANPATH=usr/share/man/man1/

# Installation path for the binary files
PATHINSTBIN=$(DESTDIR)/$(BINPATH)

# Installation path for the configuration files
PATHINSTCFG=$(DESTDIR)/$(CONFIGPATH)

# Installation path for the init file
PATHINSTINIT=$(DESTDIR)/$(INITPATH)

# Installation path for documentation
PATHINSTDOC=$(DESTDIR)/$(DOCPATH)

# Installation path for man pages
PATHINSTMAN=$(DESTDIR)/$(MANPATH)

# RPM Packaging path (where RPMs will be stored)
PATHRPMPKG=$(CURRENTDIR)/target/RPM

# DEB Packaging path (where DEBs will be stored)
PATHDEBPKG=$(CURRENTDIR)/target/DEB

# BZ2 Packaging path (where BZ2s will be stored)
PATHBZ2PKG=$(CURRENTDIR)/target/BZ2

# DOCKER Packaging path (where BZ2s will be stored)
PATHDOCKERPKG=$(CURRENTDIR)/target/DOCKER

# docker image name for consul (used during testing)
CONSUL_DOCKER_IMAGE_NAME=consul_$(VENDOR)_$(PROJECT)$(DOCKERSUFFIX)

# docker image name for NATS (used during testing)
NATS_DOCKER_IMAGE_NAME=nats_$(VENDOR)_$(PROJECT)$(DOCKERSUFFIX)

# STATIC is a flag to indicate whether to build using static or dynamic linking
ifeq ($(STATIC),0)
	STATIC_TAG=dynamic
	STATIC_FLAG=
else
	STATIC_TAG=static
	STATIC_FLAG=-static
endif

# --- MAKE TARGETS ---

# Display general help about this command
.PHONY: help
help:
	@echo ""
	@echo "$(PROJECT) Makefile."
	@echo "GOPATH=$(GOPATH)"
	@echo "The following commands are available:"
	@echo ""
	@echo "    make qa          : Run all the tests"
	@echo "    make test        : Run the unit tests"
	@echo ""
	@echo "    make format      : Format the source code"
	@echo "    make fmtcheck    : Check if the source code has been formatted"
	@echo "    make vet         : Check for syntax errors"
	@echo "    make lint        : Check for style errors"
	@echo "    make coverage    : Generate the coverage report"
	@echo "    make cyclo       : Generate the cyclomatic complexity report"
	@echo "    make ineffassign : Detect ineffectual assignments"
	@echo "    make misspell    : Detect commonly misspelled words in source files"
	@echo "    make structcheck : Find unused struct fields"
	@echo "    make varcheck    : Find unused global variables and constants"
	@echo "    make errcheck    : Check that error return values are used"
	@echo "    make staticcheck : Suggest code simplifications"
	@echo "    make astscan     : GO AST scanner"
	@echo ""
	@echo "    make docs        : Generate source code documentation"
	@echo ""
	@echo "    make deps        : Get the dependencies"
	@echo "    make build       : Compile the application"
	@echo "    make clean       : Remove any build artifact"
	@echo "    make nuke        : Deletes any intermediate file"
	@echo "    make install     : Install this application"
	@echo ""
	@echo "    make rpm         : Build an RPM package"
	@echo "    make deb         : Build a DEB package"
	@echo "    make bz2         : Build a tar bz2 (tbz2) compressed archive"
	@echo "    make docker      : Build a docker container to run this service"
	@echo "    make dockertest  : Test the newly built docker container"
	@echo ""
	@echo "    make buildall    : full build and test sequence"
	@echo "    make dbuild      : build everything inside a Docker container"
	@echo ""

# Alias for help target
all: help

# Run the unit tests (also run the NATS server)
.PHONY: test
test:
	@mkdir -p target/test
	@mkdir -p target/report
	nohup gnatsd --debug --trace > target/nats.log 2>&1 & echo $$! > target/nats.pid
	GOPATH=$(GOPATH) \
	go test \
	-tags ${STATIC_TAG} \
	-covermode=atomic \
	-bench=. \
	-race \
	-cpuprofile=target/report/cpu.out \
	-memprofile=target/report/mem.out \
	-mutexprofile=target/report/mutex.out \
	-coverprofile=target/report/coverage.out \
	-v ./src | \
	tee >(PATH=$(GOPATH)/bin:$(PATH) go-junit-report > target/test/report.xml); \
	test $${PIPESTATUS[0]} -eq 0 ; \
	echo $$? > target/test.exit; \
	kill -9 `cat target/nats.pid` ; \
	exit `cat target/test.exit`

# Format the source code
.PHONY: format
format:
	@find ./src -type f -name "*.go" -exec gofmt -s -w {} \;

# Check if the source code has been formatted
.PHONY: fmtcheck
fmtcheck:
	@mkdir -p target
	@find ./src -type f -name "*.go" -exec gofmt -s -d {} \; | tee target/format.diff
	@test ! -s target/format.diff || { echo "ERROR: the source code has not been formatted - please use 'make format' or 'gofmt'"; exit 1; }

# Validate JSON configuration files against the JSON schema
.PHONY: confcheck
confcheck:
	json validate --schema-file=resources/etc/natstest/config.schema.json --document-file=resources/test/etc/natstest/config.json
	json validate --schema-file=resources/etc/natstest/config.schema.json --document-file=resources/etc/natstest/config.json

# Check for syntax errors
.PHONY: vet
vet:
	GOPATH=$(GOPATH) \
	go vet ./src

# Check for style errors
.PHONY: lint
lint:
	GOPATH=$(GOPATH) PATH=$(GOPATH)/bin:$(PATH) golint ./src

# Generate the coverage report
.PHONY: coverage
coverage:
	@mkdir -p target/report
	GOPATH=$(GOPATH) \
	go tool cover -html=target/report/coverage.out -o target/report/coverage.html

# Report cyclomatic complexity
.PHONY: cyclo
cyclo:
	@mkdir -p target/report
	GOPATH=$(GOPATH) gocyclo -avg ./src | tee target/report/cyclo.txt ; test $${PIPESTATUS[0]} -eq 0

# Detect ineffectual assignments
.PHONY: ineffassign
ineffassign:
	@mkdir -p target/report
	GOPATH=$(GOPATH) ineffassign ./src | tee target/report/ineffassign.txt ; test $${PIPESTATUS[0]} -eq 0

# Detect commonly misspelled words in source files
.PHONY: misspell
misspell:
	@mkdir -p target/report
	GOPATH=$(GOPATH) misspell -error ./src  | tee target/report/misspell.txt ; test $${PIPESTATUS[0]} -eq 0

# Find unused struct fields.
.PHONY: structcheck
structcheck:
	@mkdir -p target/report
	GOPATH=$(GOPATH) structcheck -a ./src  | tee target/report/structcheck.txt

# Find unused global variables and constants.
.PHONY: varcheck
varcheck:
	@mkdir -p target/report
	GOPATH=$(GOPATH) varcheck -e ./src  | tee target/report/varcheck.txt

# Check that error return values are used.
.PHONY: errcheck
errcheck:
	@mkdir -p target/report
	GOPATH=$(GOPATH) errcheck ./src  | tee target/report/errcheck.txt

# Suggest code simplifications"
.PHONY: staticcheck
staticcheck:
	@mkdir -p target/report
	GOPATH=$(GOPATH) staticcheck ./src  | tee target/report/staticcheck.txt

# AST scanner
.PHONY: astscan
astscan:
	@mkdir -p target/report
	GOPATH=$(GOPATH) gosec ./src/... | tee target/report/astscan.txt ; test $${PIPESTATUS[0]} -eq 0 || true

# Generate source docs
.PHONY: docs
docs:
	@mkdir -p target/docs
	nohup sh -c 'GOPATH=$(GOPATH) godoc -http=127.0.0.1:6060' > target/godoc_server.log 2>&1 &
	wget --directory-prefix=target/docs/ --execute robots=off --retry-connrefused --recursive --no-parent --adjust-extension --page-requisites --convert-links http://127.0.0.1:6060/pkg/github.com/${VENDOR}/${PROJECT}/ ; kill -9 `lsof -ti :6060`
	@echo '<html><head><meta http-equiv="refresh" content="0;./127.0.0.1:6060/pkg/'${CVSPATH}'/'${PROJECT}'/index.html"/></head><a href="./127.0.0.1:6060/pkg/'${CVSPATH}'/'${PROJECT}'/index.html">'${PKGNAME}' Documentation ...</a></html>' > target/docs/index.html

# Alias to run targets: fmtcheck test vet lint coverage
.PHONY: qa
qa: fmtcheck confcheck test vet lint coverage cyclo ineffassign misspell structcheck varcheck errcheck staticcheck astscan

# --- INSTALL ---

# Get the dependencies
.PHONY: deps
deps:
	GOPATH=$(GOPATH) go get -tags ${STATIC_TAG} -v ./src
	GOPATH=$(GOPATH) go get github.com/nats-io/gnatsd
	GOPATH=$(GOPATH) go get github.com/inconshreveable/mousetrap
	GOPATH=$(GOPATH) go get golang.org/x/lint/golint
	GOPATH=$(GOPATH) go get github.com/jstemmer/go-junit-report
	GOPATH=$(GOPATH) go get github.com/axw/gocov/gocov
	GOPATH=$(GOPATH) go get github.com/fzipp/gocyclo
	GOPATH=$(GOPATH) go get github.com/gordonklaus/ineffassign
	GOPATH=$(GOPATH) go get github.com/client9/misspell/cmd/misspell
	GOPATH=$(GOPATH) go get github.com/opennota/check/cmd/structcheck
	GOPATH=$(GOPATH) go get github.com/opennota/check/cmd/varcheck
	GOPATH=$(GOPATH) go get github.com/kisielk/errcheck
	GOPATH=$(GOPATH) go get honnef.co/go/tools/cmd/staticcheck
	GOPATH=$(GOPATH) go get github.com/securego/gosec/cmd/gosec/...

# Install this application
.PHONY: install
install: uninstall
	mkdir -p $(PATHINSTBIN)
	cp -r ./resources/${BINPATH}* $(PATHINSTBIN)
	cp -r ./target/${BINPATH}* $(PATHINSTBIN)
	find $(PATHINSTBIN) -type d -exec chmod 755 {} \;
	find $(PATHINSTBIN) -type f -exec chmod 755 {} \;
	mkdir -p $(PATHINSTDOC)
	cp -f ./LICENSE $(PATHINSTDOC)
	cp -f ./README.md $(PATHINSTDOC)
	cp -f ./VERSION $(PATHINSTDOC)
	cp -f ./RELEASE $(PATHINSTDOC)
	chmod -R 644 $(PATHINSTDOC)*
ifneq ($(strip $(INITPATH)),)
	mkdir -p $(PATHINSTINIT)
	cp -ru ./resources/${INITPATH}* $(PATHINSTINIT)
	find $(PATHINSTINIT) -type d -exec chmod 755 {} \;
	find $(PATHINSTINIT) -type f -exec chmod 755 {} \;
endif
ifneq ($(strip $(CONFIGPATH)),)
	mkdir -p $(PATHINSTCFG)
	touch -c $(PATHINSTCFG)*
	cp -ru ./resources/${CONFIGPATH}* $(PATHINSTCFG)
	find $(PATHINSTCFG) -type d -exec chmod 755 {} \;
	find $(PATHINSTCFG) -type f -exec chmod 644 {} \;
endif
ifneq ($(strip $(MANPATH)),)
	mkdir -p $(PATHINSTMAN)
	cat ./resources/${MANPATH}${PROJECT}.1 | gzip -9 > $(PATHINSTMAN)${PROJECT}.1.gz
	find $(PATHINSTMAN) -type f -exec chmod 644 {} \;
endif

# Remove all installed files (excluding configuration files)
.PHONY: uninstall
uninstall:
	rm -rf $(PATHINSTBIN)$(PROJECT)
	rm -rf $(PATHINSTDOC)

# Remove any build artifact
.PHONY: clean
clean:
	GOPATH=$(GOPATH) go clean ./...

# Deletes any intermediate file
.PHONY: nuke
nuke:
	rm -rf ./target
	GOPATH=$(GOPATH) go clean -i ./...

# Compile the application
.PHONY: build
build: deps
	GOPATH=$(GOPATH) \
	CGO_ENABLED=0 \
	go build -tags ${STATIC_TAG} -ldflags '-linkmode external -extldflags ${STATIC_FLAG} -w -s -X main.ProgramVersion=${VERSION} -X main.ProgramRelease=${RELEASE}' -o ./target/${BINPATH}$(PROJECT) ./src
ifneq (${UPXENABLED},)
	upx --brute ./target/${BINPATH}$(PROJECT)
endif

# --- PACKAGING ---

# Build the RPM package for RedHat-like Linux distributions
.PHONY: rpm
rpm:
	rm -rf $(PATHRPMPKG)
	rpmbuild \
	--define "_topdir $(PATHRPMPKG)" \
	--define "_vendor $(VENDOR)" \
	--define "_owner $(OWNER)" \
	--define "_project $(PROJECT)" \
	--define "_package $(PKGNAME)" \
	--define "_version $(VERSION)" \
	--define "_release $(RELEASE)" \
	--define "_current_directory $(CURRENTDIR)" \
	--define "_binpath /$(BINPATH)" \
	--define "_docpath /$(DOCPATH)" \
	--define "_configpath /$(CONFIGPATH)" \
	--define "_initpath /$(INITPATH)" \
	--define "_manpath /$(MANPATH)" \
	-bb resources/rpm/rpm.spec

# Build the DEB package for Debian-like Linux distributions
.PHONY: deb
deb:
	rm -rf $(PATHDEBPKG)
	make install DESTDIR=$(PATHDEBPKG)/$(PKGNAME)-$(VERSION)
	rm -f $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/$(DOCPATH)LICENSE
	tar -zcvf $(PATHDEBPKG)/$(PKGNAME)_$(VERSION).orig.tar.gz -C $(PATHDEBPKG)/ $(PKGNAME)-$(VERSION)
	cp -rf ./resources/debian $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian
	mkdir -p $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/missing-sources
	echo "// fake source for lintian" > $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/missing-sources/$(PROJECT).c
	find $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/ -type f -exec sed -i "s/~#DATE#~/`date -R`/" {} \;
	find $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/ -type f -exec sed -i "s/~#VENDOR#~/$(VENDOR)/" {} \;
	find $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/ -type f -exec sed -i "s/~#PROJECT#~/$(PROJECT)/" {} \;
	find $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/ -type f -exec sed -i "s/~#PKGNAME#~/$(PKGNAME)/" {} \;
	find $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/ -type f -exec sed -i "s/~#VERSION#~/$(VERSION)/" {} \;
	find $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/ -type f -exec sed -i "s/~#RELEASE#~/$(RELEASE)/" {} \;
	echo $(BINPATH) > $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).dirs
	echo "$(BINPATH)* $(BINPATH)" > $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/install
	echo $(DOCPATH) >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).dirs
	echo "$(DOCPATH)* $(DOCPATH)" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/install
ifneq ($(strip $(INITPATH)),)
	echo $(INITPATH) >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).dirs
	echo "$(INITPATH)* $(INITPATH)" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/install
endif
ifneq ($(strip $(CONFIGPATH)),)
	echo $(CONFIGPATH) >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).dirs
	echo "$(CONFIGPATH)* $(CONFIGPATH)" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/install
endif
ifneq ($(strip $(MANPATH)),)
	echo $(MANPATH) >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).dirs
	echo "$(MANPATH)* $(MANPATH)" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/install
endif
	echo "new-package-should-close-itp-bug" > $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).lintian-overrides
	echo "hardening-no-relro $(BINPATH)$(PROJECT)" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).lintian-overrides
	echo "embedded-library $(BINPATH)$(PROJECT): libyaml" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).lintian-overrides
	echo "statically-linked-binary usr/bin/natstest" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).lintian-overrides
	echo "script-with-language-extension usr/bin/md5str.sh" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).lintian-overrides
	echo "binary-without-manpage usr/bin/md5str.sh" >> $(PATHDEBPKG)/$(PKGNAME)-$(VERSION)/debian/$(PKGNAME).lintian-overrides
	cd $(PATHDEBPKG)/$(PKGNAME)-$(VERSION) && debuild -us -uc

# build a compressed bz2 archive
.PHONY: bz2
bz2:
	rm -rf $(PATHBZ2PKG)
	make install DESTDIR=$(PATHBZ2PKG)
	tar -jcvf $(PATHBZ2PKG)/$(PKGNAME)-$(VERSION)-$(RELEASE).tbz2 -C $(PATHBZ2PKG) usr/ etc/

# build a docker container to run this service
.PHONY: docker
docker:
	rm -rf $(PATHDOCKERPKG)
	make install DESTDIR=$(PATHDOCKERPKG)
	cp resources/DockerDeploy/Dockerfile $(PATHDOCKERPKG)/
	docker build --no-cache --tag=$(VENDOR)/$(PROJECT)$(DOCKERSUFFIX):latest $(PATHDOCKERPKG)

# check if the deployment container starts
.PHONY: dockertest
dockertest:
	# clean previous docker containers (if any)
	rm -f target/old_docker_containers.id
	docker ps -a | grep $(NATS_DOCKER_IMAGE_NAME) | awk '{print $$1}' >> target/old_docker_containers.id || true
	docker ps -a | grep $(CONSUL_DOCKER_IMAGE_NAME) | awk '{print $$1}' >> target/old_docker_containers.id || true
	docker ps -a | grep $(VENDOR)/$(PROJECT)$(DOCKERSUFFIX) | awk '{print $$1}' >> target/old_docker_containers.id || true
	docker stop `cat target/old_docker_containers.id` 2> /dev/null || true
	docker rm `cat target/old_docker_containers.id` 2> /dev/null || true
	# start a NATS service inside a container
	docker run --detach=true --name=$(NATS_DOCKER_IMAGE_NAME)_$(VERSION)-$(RELEASE) --publish=4222 --hostname=test.nats nats > target/nats_docker_container.id
	# start a Consul service inside a container
	docker run --detach=true --name=$(CONSUL_DOCKER_IMAGE_NAME)_$(VERSION)-$(RELEASE) --publish=8500 --hostname=test.consul progrium/consul -server -bootstrap > target/consul_docker_container.id
	sleep 5
	# Get Docker ports
	docker inspect --format='{{(index (index .NetworkSettings.Ports "4222/tcp") 0).HostPort}}' `cat target/nats_docker_container.id` > target/nats_docker_container.port
	docker inspect --format='{{(index (index .NetworkSettings.Ports "8500/tcp") 0).HostPort}}' `cat target/consul_docker_container.id` > target/consul_docker_container.port
	# push Consul configuration
	curl --request PUT --data @resources/test/etc/natstest/config.json http://127.0.0.1:`cat target/consul_docker_container.port`/v1/kv/config/natstest
	# Start natstest container
	docker run --detach=true --net="host" --tty=true \
	--env="NATSTEST_REMOTECONFIGPROVIDER=consul" \
	--env="NATSTEST_REMOTECONFIGENDPOINT=127.0.0.1:`cat target/consul_docker_container.port`" \
	--env="NATSTEST_REMOTECONFIGPATH=/config/natstest" \
	--env="NATSTEST_REMOTECONFIGSECRETKEYRING=" \
	${VENDOR}/${PROJECT}$(DOCKERSUFFIX):latest > target/project_docker_container.id || true
	sleep 5
	# check if the container is working
	docker inspect -f {{.State.Running}} `cat target/project_docker_container.id` > target/project_docker_container.run || true
	# remove the testing container
	docker stop `cat target/project_docker_container.id` 2> /dev/null || true
	docker logs `cat target/project_docker_container.id` || true
	docker rm `cat target/project_docker_container.id` 2> /dev/null || true
	docker stop `cat target/consul_docker_container.id` 2> /dev/null || true
	docker rm `cat target/consul_docker_container.id` 2> /dev/null || true
	docker stop `cat target/nats_docker_container.id` 2> /dev/null || true
	docker rm `cat target/nats_docker_container.id` 2> /dev/null || true
	@exit `grep -ic "false" target/project_docker_container.run`

# Full build and test sequence
.PHONY: buildall
buildall: build qa rpm deb

# Build everything inside a Docker container
.PHONY: dbuild
dbuild:
	@mkdir -p target
	@rm -rf target/*
	@echo 0 > target/make.exit
	CVSPATH=$(CVSPATH) VENDOR=$(VENDOR) PROJECT=$(PROJECT) MAKETARGET='$(MAKETARGET)' ./dockerbuild.sh
	@exit `cat target/make.exit`
