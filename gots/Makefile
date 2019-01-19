all:
	go build
	./gots -help

arm:
	GOOS=linux GOARCH=arm GOARM=7 go build

win:
	GOOS=windows GOARCH=amd64 go build
