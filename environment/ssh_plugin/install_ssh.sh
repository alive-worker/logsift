#!/bin/sh
# Install and configure OpenSSH server inside the image.
set -eu

SSH_PASSWORD="${SSH_PASSWORD:-password}"
SSH_PORT="${SSH_PORT:-22}"

log() { printf '[ssh_plugin/install] %s\n' "$*"; }

if command -v sshd >/dev/null 2>&1; then
    log "sshd already present, skipping package installation"
else
    if command -v apt-get >/dev/null 2>&1; then
        export DEBIAN_FRONTEND=noninteractive
        apt-get update
        apt-get install -y --no-install-recommends openssh-server ca-certificates wget curl
        rm -rf /var/lib/apt/lists/*
    elif command -v apk >/dev/null 2>&1; then
        apk add --no-cache openssh openssh-keygen shadow ca-certificates wget curl
    elif command -v dnf >/dev/null 2>&1; then
        dnf install -y openssh-server openssh-clients wget curl
        dnf clean all
    elif command -v yum >/dev/null 2>&1; then
        yum install -y openssh-server openssh-clients wget curl
        yum clean all
    elif command -v zypper >/dev/null 2>&1; then
        zypper --non-interactive install openssh wget curl
        zypper clean -a || true
    else
        log "ERROR: no supported package manager found" >&2
        exit 1
    fi
fi

mkdir -p /var/run/sshd /root/.ssh
chmod 700 /root/.ssh

if command -v chpasswd >/dev/null 2>&1; then
    printf 'root:%s\n' "${SSH_PASSWORD}" | chpasswd
elif command -v passwd >/dev/null 2>&1; then
    printf '%s\n%s\n' "${SSH_PASSWORD}" "${SSH_PASSWORD}" | passwd root >/dev/null
else
    log "WARNING: neither chpasswd nor passwd available; root password not set" >&2
fi

mkdir -p /etc/ssh/sshd_config.d
cat > /etc/ssh/sshd_config.d/99-ssh-plugin.conf <<EOF
# Managed by ssh_plugin/install_ssh.sh.
Port ${SSH_PORT}
PermitRootLogin yes
PasswordAuthentication yes
PubkeyAuthentication yes
UsePAM no
EOF

chmod 644 /etc/ssh/sshd_config.d/99-ssh-plugin.conf

if [ -f /etc/ssh/sshd_config ]; then
    if ! grep -qE '^[[:space:]]*Include[[:space:]]+/etc/ssh/sshd_config\.d/' /etc/ssh/sshd_config; then
        printf '\n# added by ssh_plugin\nInclude /etc/ssh/sshd_config.d/*.conf\n' >> /etc/ssh/sshd_config
    fi
fi

ssh-keygen -A >/dev/null 2>&1 || true
log "done: openssh-server ready; root password set, listening on port ${SSH_PORT}"
