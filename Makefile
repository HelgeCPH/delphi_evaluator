.PHONY: build clean

build:
	GOOS=linux GOARCH=386 go build -o build/phase_one_server phase_one_server.go
	GOOS=linux GOARCH=386 go build -o build/phase_two_server phase_two_server.go
	GOOS=windows GOARCH=386 go build -o build/phase_one_server.exe phase_one_server.go
	GOOS=windows GOARCH=386 go build -o build/phase_two_server.exe phase_two_server.go
	GOOS=darwin GOARCH=386 go build -o build/phase_one_server_mac phase_one_server.go
	GOOS=darwin GOARCH=386 go build -o build/phase_two_server_mac phase_two_server.go


clean:
	rm build/phase_*
