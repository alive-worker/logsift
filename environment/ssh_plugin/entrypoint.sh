#!/bin/sh
# Container entrypoint for SSH-first development images.
set -u

KEEP_ALIVE="${KEEP_ALIVE:-1}"
SKIP_CMD="${SKIP_CMD:-0}"
ORIG_ENTRYPOINT="${ORIG_ENTRYPOINT:-}"

log() { printf '[ssh_plugin/entrypoint] %s\n' "$*"; }

if [ ! -f /etc/ssh/ssh_host_rsa_key ] && [ ! -f /etc/ssh/ssh_host_ed25519_key ]; then
    ssh-keygen -A >/dev/null 2>&1 || true
fi

mkdir -p /var/run/sshd
/usr/sbin/sshd

actual_port=$(
    { cat /etc/ssh/sshd_config.d/*.conf 2>/dev/null; cat /etc/ssh/sshd_config 2>/dev/null; } \
    | awk 'BEGIN{p=22} /^[[:space:]]*Port[[:space:]]+[0-9]+/ {p=$2} END{print p}'
)
log "sshd started on port ${actual_port}"

have_cmd=0
if [ "$#" -gt 0 ] && [ "$SKIP_CMD" != "1" ]; then
    have_cmd=1
fi

run_business() {
    if [ -n "$ORIG_ENTRYPOINT" ] && [ -x "$ORIG_ENTRYPOINT" ]; then
        "$ORIG_ENTRYPOINT" "$@"
    else
        "$@"
    fi
}

if [ "$KEEP_ALIVE" = "0" ]; then
    if [ "$have_cmd" -eq 0 ]; then
        log "KEEP_ALIVE=0 but no CMD given; falling back to idle wait"
        exec tail -f /dev/null
    fi

    log "KEEP_ALIVE=0: exec business CMD: $*"
    if [ -n "$ORIG_ENTRYPOINT" ] && [ -x "$ORIG_ENTRYPOINT" ]; then
        exec "$ORIG_ENTRYPOINT" "$@"
    else
        exec "$@"
    fi
fi

child_pid=""
cleanup() {
    log "received signal, shutting down ..."
    if [ -f /var/run/sshd.pid ]; then
        kill -TERM "$(cat /var/run/sshd.pid)" 2>/dev/null || true
    fi
    if [ -n "$child_pid" ]; then
        kill -TERM "$child_pid" 2>/dev/null || true
        wait "$child_pid" 2>/dev/null || true
    fi
    exit 0
}
trap cleanup TERM INT HUP

if [ "$have_cmd" -eq 1 ]; then
    log "running business CMD in background: $*"
    (
        run_business "$@"
        rc=$?
        printf '[ssh_plugin/entrypoint] business CMD exited with code %s; SSH remains available\n' "$rc"
    ) &
    child_pid=$!
else
    log "no business CMD (SSH-only mode)"
fi

while :; do
    wait 2>/dev/null || true
    sleep 3600 &
    wait "$!" 2>/dev/null || true
done
