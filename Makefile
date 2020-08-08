build:
	docker run -it --rm \
		-v`pwd`:/go/src/github.com/velp/mass \
		-w /go/src/github.com/velp/mass golang:1.14 \
		/bin/bash -c "apt-get update -qq && \
					  apt-get install libpcap-dev build-essential -y && \
					  go mod download && \
					  CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o mass"
