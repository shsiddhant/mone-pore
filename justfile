app := "mone-pore"
version_pkg := "github.com/shsiddhant/mone-pore"

version := `git describe --tags --always --dirty 2>/dev/null || echo 0.0.0`

# Build binary for local machine
build:
    go build -ldflags="-X {{version_pkg}}.Version={{version}}" -o bin/{{app}}-{{version}} .

# Build binary for pi (64-bit)
build-pi:
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-X {{version_pkg}}.Version={{version}}" -o bin/{{app}}-{{version}}-linux-arm64 .

# Clean generated binaries
clean:
    rm -rf bin/
