.PHONY: run-all
run-all: setup run-test clean

NSE=1
.PHONY: setup
setup:
	NSE=${NSE} ./scripts/setup.sh


# Enviroment varibles for go test
LOG=on
TIMEOUT=300
INIT=on
NSE_LOG=30
NSMGR_LOG=30
PING_LOG=on
# run test again after setup
.PHONY: run-test
run-test:
	INIT=${INIT} LOG=${LOG} TIMEOUT=${TIMEOUT} NSE_LOG=${NSE_LOG} NSMGR_LOG=${NSMGR_LOG} PING_LOG=${PING_LOG} go test

.PHONY: clean
clean:
	./scripts/clean.sh