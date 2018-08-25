# Copyright 2017 Gavin Chun Jin. All rights reserved.
# Use of this source code is governed by the MIT
# license that can be found in the LICENSE file.

.PHONY: all setcapabilities benchmark clean fmt test gometa

all: setcapabilities

clean:
	test -n "$(GOBIN)"  # $$GOBIN
	@echo ... using GOBIN $$GOBIN
	rm ${GOBIN}/goping

test:
	GOCACHE=off go test -cpu=1,2 ./core

fmt:
	go fmt .
	go fmt ./core
	go fmt ./thirdparty

gometa:
	test -n "$(GOBIN)"  # $$GOBIN
	${GOBIN}/gometalinter --disable-all --enable=errcheck --enable=vet --enable=vetshadow ./...

benchmark:
	go test ./core -bench .

${GOBIN}/goping: fmt test
	go install

setcapabilities: ${GOBIN}/goping
	@echo ... using GOBIN $$GOBIN
	sudo /sbin/setcap cap_net_raw+ep ${GOBIN}/goping
