NSE=1

.PHONY: run-all
run-all: setup clean

.PHONY: setup
setup:
	NSE=${NSE} ./scripts/setup.sh

INIT=off
LOG=on
TIMEOUT=300
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