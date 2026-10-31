#!/usr/bin/env bash
set -euo pipefail
cd /opt/yourmove
umask 077
mkdir -p backups
destination="backups/yourmove-$(date -u +%Y%m%dT%H%M%SZ).dump"
trap 'rm -f "$destination.partial"' EXIT
docker compose exec -T db pg_dump -U postgres -d yourmove --format=custom > "$destination.partial"
docker compose exec -T db pg_restore --list < "$destination.partial" > /dev/null
mv "$destination.partial" "$destination"
find backups -maxdepth 1 -name 'yourmove-*.dump' -type f -mtime +14 -delete
