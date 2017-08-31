default: build

# build generate binary on './bin' directory.
build: 
	go build -o bin/exporter_proxy .
