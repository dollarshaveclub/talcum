.PHONY: talcum

talcum:
	go install github.com/dollarshaveclub/talcum/src/cmd/talcum

release:
	rm -rf build/*
	GOOS=linux GOARCH=amd64 go build -o build/talcum github.com/dollarshaveclub/talcum/src/cmd/talcum
	tar -c -C build talcum | gzip -c > build/talcum_linux_amd64.tar.gz
