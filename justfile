app := "mone-pore"
version_pkg := "github.com/shsiddhant/mone-pore"

version := `git describe --tags --always --dirty 2>/dev/null || echo 0.0.0`

pi_host := "pi"

# Build binary for local machine
build:
    go build -ldflags="-X {{version_pkg}}.Version={{version}}" -o bin/{{app}}-{{version}} .

# Build binary for pi (64-bit)
build-pi:
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-X {{version_pkg}}.Version={{version}}" -o bin/{{app}}-{{version}}-linux-arm64 .

# Build and deploy to Raspberry Pi, then restart the service
install-pi: build-pi
    @echo "Deploying {{app}} version {{version}} to Pi..."
    cat bin/{{app}}-{{version}}-linux-arm64 | ssh {{pi_host}} "install -m 755 /dev/stdin ~/.local/bin/{{app}}"
    @echo "Restarting systemd service..."
    ssh {{pi_host}} "systemctl --user restart {{app}}.service"
    @echo "Deployment complete"

# Clean generated binaries
clean:
    rm -rf bin/
