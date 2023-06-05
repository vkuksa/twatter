# Change into parent directory
CWD="$( cd "$( dirname "$0" )"/.. && pwd )"
echo "Working directory: $CWD"
cd "$CWD"

# Delete the old dir
echo "==> Removing old directory..."
rm -f bin/*
rm -rf pkg/*
mkdir -p bin/

# Instruct gox to build statically linked binaries
export CGO_ENABLED=0

# Set module download mode to readonly to not implicitly update go.mod
export GOFLAGS="-mod=readonly"

echo "==> Downloading modules..."

# Ensure all remote modules are downloaded and cached before build so that
# the concurrent builds launched by gox won't race to redundantly download them.
go mod download

echo "==> Building..."

GOOS=linux GOARCH=amd64 go build -o bin/twatterd cmd/twatterd/main.go 
GOOS=linux GOARCH=amd64 go build -o bin/spammer cmd/spammer/main.go 
