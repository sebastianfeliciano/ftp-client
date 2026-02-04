.PHONY: build clean

build:
	go build -o 4700ftp .

clean:
	rm -f 4700ftp
