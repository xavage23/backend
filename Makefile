all:
	CGO_ENABLED=0 go build -v 
	systemctl reload stocksim2-api
tests:
	CGO_ENABLED=0 go test -v -coverprofile=coverage.out ./...
ts:
	~/go/bin/tygo generate
