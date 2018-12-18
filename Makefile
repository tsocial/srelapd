# Make dependencies

.PHONY: build

SERVER_REPO := "tsl8/srelapd"

SHELL = /bin/bash

deps:
	dep version || (curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh)
	dep ensure -v

build: deps
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o dist/srelapd -a -installsuffix cgo .

docker_build: build
	docker-compose -f build/docker-compose.yaml build srelapd

docker_upload:
	docker tag $(SERVER_REPO):latest $(SERVER_REPO):$(TRAVIS_BRANCH)-latest
	docker tag $(SERVER_REPO):latest $(SERVER_REPO):$(TRAVIS_BRANCH)-$(TRAVIS_BUILD_NUMBER)
	docker push $(SERVER_REPO):latest
	docker push $(SERVER_REPO):$(TRAVIS_BRANCH)-latest
	docker push $(SERVER_REPO):$(TRAVIS_BRANCH)-$(TRAVIS_BUILD_NUMBER)
