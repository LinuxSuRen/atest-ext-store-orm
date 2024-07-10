fmt:
	go mod tidy
	go fmt ./...
build:
	go build -o bin/atest-store-orm .
test:
	go test ./... -cover -v -coverprofile=coverage.out
	go tool cover -func=coverage.out
build-image:
	docker build .
hd:
	curl https://linuxsuren.github.io/tools/install.sh|bash
init-env: hd
	hd i cli/cli
	gh extension install linuxsuren/gh-dev
run-e2e:
	cd e2e && ./start.sh
