@default: run

@run:
  go run .

test:
  go test -v ./...

clean:
  rm -rf build/
  rm -rf ~/.cache/communities-tui/*
  rm -rf ~/.local/state/communities-tui/*

build:
  @echo "Building communities-tui..."

  @echo "Creating build directory at build/..."
  mkdir -p build

  @echo "Installing dependencies..."
  go mod tidy

  @echo "Building communities-tui application..."
  go build -o build/communities-tui main.go

  @echo "Build complete."

install: build
  @echo "Installing communities-tui..."
  ./install.sh
  @echo "Installation complete."

uninstall: clean
  @echo "Cleaning communities-tui..."
  sudo rm -f /usr/local/bin/communities-tui
  @echo "Clean complete"
