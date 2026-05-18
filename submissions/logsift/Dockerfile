# Single-stage image for the trial-labeling initial scene.
# Spec §3.2: WORKDIR /app, git-initialised, minimal usable init.
FROM golang:1.24

# Required by spec: container repo path is /app.
WORKDIR /app

# Use a China-friendly Go module proxy so `go mod download` is reliable in
# mainland network conditions; the proxy is upstream-mirrored, package
# contents are identical.
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
# Ensure `go` is on PATH for non-login shells too.
ENV PATH=/usr/local/go/bin:/go/bin:$PATH

# System deps. git is mandated by spec; tmux/curl help interactive sessions.
RUN set -eux; \
    sed -i 's|http://deb.debian.org|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || true; \
    apt-get update && apt-get install -y --no-install-recommends \
        git curl tmux ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Pre-cache Go modules so the first `go test` inside the container is fast.
COPY go.mod ./
RUN go mod download

# Copy the rest of the source.
COPY . ./

# Ensure /app is a clean git repo at the task starting state.
RUN git init -q -b main \
    && git config user.email "agent@logsift.local" \
    && git config user.name  "logsift-agent" \
    && git add -A \
    && git commit -q -m "initial scene"

CMD ["/bin/bash"]
