.PHONY: docker
docker:
	@if exist onlinejudge del /f onlinejudge
	@set CGO_ENABLED=0
	@set GOOS=linux
	@set GOARCH=amd64
	@go build -tags=k8s -o onlinejudge .
	@docker rmi -f crazyfran/onlinejudge:v0.0.1 .
	@docker build -t crazyfran/onlinejudge:v0.0.1 .