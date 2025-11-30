#!/bin/bash
if [ -z "$1" ]; then
    echo "Error: No Go version parameter provided."
    exit 1
fi

GO_VERSION="$1"
GOLANGCILINT_VERSION="${2:-"latest"}"

set -eux 
apt-get update && apt-get dist-upgrade -y
apt-get install -y ca-certificates curl gcc git nano procps sudo tzdata wget zsh build-essential

 usermod -aG sudo vscode 
 echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

TARGET_GOROOT="${TARGET_GOROOT:-"/usr/local/go"}"

# Install Go tools that are isImportant && !replacedByGopls based on
# https://github.com/golang/vscode-go/blob/v0.38.0/src/goToolsInformation.ts
GO_TOOLS="\
    golang.org/x/tools/gopls@latest \
    honnef.co/go/tools/cmd/staticcheck@latest \
    golang.org/x/lint/golint@latest \
    github.com/mgechev/revive@latest \
    github.com/go-delve/delve/cmd/dlv@latest \
    github.com/fatih/gomodifytags@latest \
    github.com/haya14busa/goplay/cmd/goplay@latest \
    github.com/cweill/gotests/gotests@latest \ r
    github.com/josharian/impl@latest \
    mvdan.cc/gofumpt@latest \
    github.com/matryer/moq@latest \
    github.com/divan/expvarmon@latest \
	github.com/rakyll/hey@latest \
	golang.org/x/vuln/cmd/govulncheck@latest \
	golang.org/x/tools/cmd/goimports@latest"

    # bitbucket.org/styletronic/dkvlive-async-api-parser@latest" \

echo "Installing common Go tools..."
export PATH=${TARGET_GOROOT}/bin:${PATH}
export GOPATH=/tmp/gotools
export GOCACHE=/tmp/gotools/cache
mkdir -p /tmp/gotools /usr/local/etc/vscode-dev-containers
cd /tmp/gotools    

(echo "${GO_TOOLS}" | xargs -n 1 go install -v )2>&1 | tee -a /usr/local/etc/vscode-dev-containers/go.log
# Move Go tools into path and clean up
if [ -d /tmp/gotools/bin ]; then
    mv /tmp/gotools/bin/* ${TARGET_GOROOT}/bin/
    rm -rf /tmp/gotools
fi

# Install golangci-lint from precompiled binares
if [ "$GOLANGCILINT_VERSION" = "latest" ] || [ "$GOLANGCILINT_VERSION" = "" ]; then
    echo "Installing golangci-lint latest..."
    curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
        sh -s -- -b "${TARGET_GOROOT}/bin"
else
    echo "Installing golangci-lint ${GOLANGCILINT_VERSION}..."
    curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
        sh -s -- -b "${TARGET_GOROOT}/bin" "${GOLANGCILINT_VERSION}"
fi

rm -rf /var/lib/apt/lists/*
# Add Go bin to bash and zsh profiles if not already present
for profile in /etc/profile /etc/bash.bashrc /etc/zsh/zshrc; do
    if [ -f "$profile" ]; then
        grep -qxF "export PATH=\"${TARGET_GOROOT}/bin:\$PATH\"" "$profile" || \
            echo "export PATH=\"${TARGET_GOROOT}/bin:\$PATH\"" >> "$profile"
    fi
done

# Add common shortcuts like 'll' for all users
for shellrc in /etc/bash.bashrc /etc/zsh/zshrc; do
    echo "alias ll='ls -alF'" >> "$shellrc"
    echo "alias la='ls -A'" >> "$shellrc"
    echo "alias l='ls -CF'" >> "$shellrc"
    echo "alias ga='ga.sh'" >> "$shellrc"
    echo "export GOPRIVATE='bitbucket.org/styletronic'" >> "$shellrc"
done

su vscode -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
echo >> /home/vscode/.bashrc
echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/vscode/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"

su vscode -c "brew install act"