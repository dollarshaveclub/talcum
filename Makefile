.PHONY: talcum

talcum:
	go install github.com/dollarshaveclub/talcum/src/cmd/talcum

release:
	rm -rf build/*
	docker run --rm -v `pwd`:/go/src/github.com/dollarshaveclub/talcum \
		-w /go/src/github.com/dollarshaveclub/talcum golang:1.7.5-alpine \
		go build -o build/talcum github.com/dollarshaveclub/talcum/src/cmd/talcum
	tar -c -C build talcum | gzip -c > build/talcum_linux_amd64.tar.gz
