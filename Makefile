.PHONY: network-up network-down build-contracts

network-up:
	cd fabric-samples/test-network && ./network.sh up createChannel -c mychannel -ca -s couchdb

network-down:
	cd fabric-samples/test-network && ./network.sh down

build-contracts:
	go build ./types/...
	go build ./chaincodes/audit/...
	go build ./chaincodes/basic-asset/...
