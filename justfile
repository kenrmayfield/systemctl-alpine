build-linux:
    mkdir -p dist
    GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o dist/systemctl-amd64 .
    GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o dist/systemctl-arm64 .
