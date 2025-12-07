@default: run

@run:
  go run .

test:
  go test -v ./...

clean:
  rm -rf build/
  rm -rf ~/.cache/tuistr/*
  rm -rf ~/.local/state/tuistr/*

build:
  @echo "Building tuistr..."

  @echo "Creating build directory at build/..."
  mkdir -p build

  @echo "Installing dependencies..."
  go mod tidy

  @echo "Building tuistr application..."
  go build -o build/tuistr main.go

  @echo "Build complete."

install: build
  @echo "Installing tuistr..."
  ./install.sh
  @echo "Installation complete."

uninstall: clean
  @echo "Cleaning tuistr..."
  sudo rm -f /usr/local/bin/tuistr
  @echo "Clean complete"
