# Copyright 2017 Gavin Bong. All rights reserved.
# Use of this source code is governed by the MIT
# license that can be found in the LICENSE file.

.PHONY: all setcapabilities benchmark clean

all: setcapabilities

clean:
	@echo ... using GOBIN $$GOBIN
	rm ${GOBIN}/goping

benchmark:
	go test ./core -bench .

${GOBIN}/goping:
	go test -cpu=1,2 ./core
	go install

setcapabilities: ${GOBIN}/goping
	@echo ... using GOBIN $$GOBIN
	sudo /sbin/setcap cap_net_raw=ep ${GOBIN}/goping
