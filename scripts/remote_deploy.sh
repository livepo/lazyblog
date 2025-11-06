#!/usr/bin/env bash
set -euo pipefail

# remote_deploy.sh
# Usage (on target server):
#   bash remote_deploy.sh --service lazyblog --target /opt/lazyblog --archive ~/deploy.tar.gz
# Or run as regular user (script uses for privileged ops):
#   bash remote_deploy.sh --service lazyblog --target /opt/lazyblog --archive ~/deploy.tar.gz
#
# Environment / options:
#   --service <name>           (required) systemd service name to stop/start
#   --target <dir>             (required) directory where lazyblog/static/templates live
#   --archive <path>           (optional, default: ~/deploy.tar.gz)
#   --backups-dir <dir>        (optional, default: /var/backups)
#   --keep <n>                 (optional, default: 3) how many backups to keep
#   --health-url <url>         (optional, default: http://127.0.0.1:8080/ready)
#   --health-timeout <secs>    (optional, default: 2)
#   --help

print_usage() {
	sed -n '1,120p' "$0" | sed -n '1,40p'
}

ARGS_ARCHIVE="~/deploy.tar.gz"
ARGS_BACKUPS_DIR="/var/backups"
ARGS_KEEP=3
ARGS_HEALTH_URL="http://127.0.0.1:8080/ready"
ARGS_HEALTH_TIMEOUT=2

while [[ $# -gt 0 ]]; do
	case "$1" in
	--service)
		SERVICE_NAME="$2"
		shift 2
		;;
	--target)
		TARGET_DIR="$2"
		shift 2
		;;
	--archive)
		ARGS_ARCHIVE="$2"
		shift 2
		;;
	--backups-dir)
		ARGS_BACKUPS_DIR="$2"
		shift 2
		;;
	--keep)
		ARGS_KEEP="$2"
		shift 2
		;;
	--health-url)
		ARGS_HEALTH_URL="$2"
		shift 2
		;;
	--health-timeout)
		ARGS_HEALTH_TIMEOUT="$2"
		shift 2
		;;
	-h | --help)
		print_usage
		exit 0
		;;
	*)
		echo "Unknown arg: $1"
		print_usage
		exit 2
		;;
	esac
done

if [[ -z "${SERVICE_NAME:-}" ]]; then
	echo "ERROR: --service is required" >&2
	exit 2
fi
if [[ -z "${TARGET_DIR:-}" ]]; then
	echo "ERROR: --target is required" >&2
	exit 2
fi

# Expand ~ in archive path
ARCHIVE_PATH="${ARGS_ARCHIVE/#~/$HOME}"
BACKUPS_DIR="$ARGS_BACKUPS_DIR"
KEEP=$ARGS_KEEP
HEALTH_URL="$ARGS_HEALTH_URL"
HEALTH_TIMEOUT=$ARGS_HEALTH_TIMEOUT

log() { printf '%s %s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$*"; }

log "Starting remote_deploy: service=$SERVICE_NAME target=$TARGET_DIR archive=$ARCHIVE_PATH"

if [[ ! -f "$ARCHIVE_PATH" ]]; then
	echo "ERROR: archive not found: $ARCHIVE_PATH" >&2
	exit 3
fi

BACKUP_DIR="$BACKUPS_DIR/${SERVICE_NAME}-$(date +%s)"
log "Creating backup dir $BACKUP_DIR"
mkdir -p "$BACKUP_DIR"

log "Stopping service $SERVICE_NAME (if running)"
systemctl stop "$SERVICE_NAME" || true

# Move existing files to backup
if [[ -f "$TARGET_DIR/lazyblog" ]]; then
	log "Backing up binary"
	mv -f "$TARGET_DIR/lazyblog" "$BACKUP_DIR/"
fi
if [[ -d "$TARGET_DIR/static" ]]; then
	log "Backing up static/"
	mv -f "$TARGET_DIR/static" "$BACKUP_DIR/"
fi
if [[ -d "$TARGET_DIR/templates" ]]; then
	log "Backing up templates/"
	mv -f "$TARGET_DIR/templates" "$BACKUP_DIR/"
fi

TMPDIR=$(mktemp -d /tmp/deploy.XXXXXX)
log "Extracting $ARCHIVE_PATH to $TMPDIR"
tar -xzf "$ARCHIVE_PATH" -C "$TMPDIR"

log "Ensuring target dir exists: $TARGET_DIR"
mkdir -p "$TARGET_DIR"

if [[ -f "$TMPDIR/lazyblog" ]]; then
	log "Installing new binary"
	mv -f "$TMPDIR/lazyblog" "$TARGET_DIR/lazyblog"
	chmod +x "$TARGET_DIR/lazyblog"
fi

if [[ -d "$TMPDIR/static" ]]; then
	log "Installing new static/"
	rm -rf "$TARGET_DIR/static"
	mv -f "$TMPDIR/static" "$TARGET_DIR/static"
fi

if [[ -d "$TMPDIR/templates" ]]; then
	log "Installing new templates/"
	rm -rf "$TARGET_DIR/templates"
	mv -f "$TMPDIR/templates" "$TARGET_DIR/templates"
fi

log "Cleaning temporary files"
rm -rf "$TMPDIR"
rm -f "$ARCHIVE_PATH"

log "Starting service $SERVICE_NAME"
systemctl start "$SERVICE_NAME"
sleep 3

log "Running health check against $HEALTH_URL"
if curl -sSf --max-time "$HEALTH_TIMEOUT" "$HEALTH_URL" >/dev/null 2>&1; then
	log "Health check passed"
	# prune old backups, keep $KEEP latest
	if [[ $KEEP -gt 0 ]]; then
		log "Pruning old backups, keeping $KEEP"
		ls -dt "$BACKUPS_DIR/${SERVICE_NAME}-*" 2>/dev/null | tail -n +$((KEEP + 1)) | xargs -r rm -rf || true
	fi
	log "Deploy succeeded"
	exit 0
else
	log "Health check FAILED, attempting rollback"
	systemctl stop "$SERVICE_NAME" || true
	if [[ -d "$BACKUP_DIR" ]]; then
		log "Restoring backup from $BACKUP_DIR"
		if [[ -f "$BACKUP_DIR/lazyblog" ]]; then
			mv -f "$BACKUP_DIR/lazyblog" "$TARGET_DIR/lazyblog" || true
			chmod +x "$TARGET_DIR/lazyblog" || true
		fi
		if [[ -d "$BACKUP_DIR/static" ]]; then
			rm -rf "$TARGET_DIR/static" || true
			mv -f "$BACKUP_DIR/static" "$TARGET_DIR/static" || true
		fi
		if [[ -d "$BACKUP_DIR/templates" ]]; then
			rm -rf "$TARGET_DIR/templates" || true
			mv -f "$BACKUP_DIR/templates" "$TARGET_DIR/templates" || true
		fi
		systemctl start "$SERVICE_NAME" || true
		log "Rollback complete; service started"
	else
		log "No backup found to rollback"
	fi
	exit 4
fi
