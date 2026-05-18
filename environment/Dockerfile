FROM golang:1.24-bookworm

WORKDIR /app

ARG http_proxy=
ARG https_proxy=
ARG HTTP_PROXY=
ARG HTTPS_PROXY=

ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    PATH=/usr/local/go/bin:/go/bin:$PATH \
    http_proxy=${http_proxy} \
    https_proxy=${https_proxy} \
    HTTP_PROXY=${HTTP_PROXY} \
    HTTPS_PROXY=${HTTPS_PROXY} \
    SSH_PORT=22 \
    SKIP_CMD=1

RUN set -eux; \
    sed -i 's|http://deb.debian.org|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || true; \
    apt-get update && apt-get install -y --no-install-recommends \
        git curl tmux ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN git init -q -b main \
    && git config user.email "agent@logsift.local" \
    && git config user.name "logsift-agent" \
    && git add -A \
    && git commit -q -m "initial scene"

USER root
COPY environment/ssh_plugin/ /opt/ssh_plugin/
RUN chmod +x /opt/ssh_plugin/*.sh && /opt/ssh_plugin/install_ssh.sh

EXPOSE 22
ENTRYPOINT ["/opt/ssh_plugin/entrypoint.sh"]
CMD ["bash"]
