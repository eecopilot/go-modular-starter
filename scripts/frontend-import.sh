#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

FRONTEND_APP_DIR=${FRONTEND_APP_DIR:-web/app}
FRONTEND_OUTPUT_DIR=${FRONTEND_OUTPUT_DIR:-web/dist}
FRONTEND_DIST_DIR=${FRONTEND_DIST_DIR:-}
FRONTEND_GIT_URL=${FRONTEND_GIT_URL:-}
FRONTEND_GIT_REF=${FRONTEND_GIT_REF:-}
FRONTEND_INSTALL_COMMAND=${FRONTEND_INSTALL_COMMAND:-}
FRONTEND_BUILD_COMMAND=${FRONTEND_BUILD_COMMAND:-}
FRONTEND_SKIP_INSTALL=${FRONTEND_SKIP_INSTALL:-false}
FRONTEND_REQUIRES_INDEX=${FRONTEND_REQUIRES_INDEX:-true}

log() {
	printf '%s\n' "$*"
}

fail() {
	printf 'frontend-import: %s\n' "$*" >&2
	exit 1
}

abs_path() {
	case "$1" in
		/*) printf '%s\n' "$1" ;;
		*) printf '%s/%s\n' "$ROOT_DIR" "$1" ;;
	esac
}

reject_parent_ref() {
	case "/$1/" in
		*/../*) fail "$2 must not contain .. path segments" ;;
	esac
}

bool_enabled() {
	case "$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')" in
		1|true|yes|y|on) return 0 ;;
		*) return 1 ;;
	esac
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

APP_DIR=$(abs_path "$FRONTEND_APP_DIR")
OUTPUT_DIR=$(abs_path "$FRONTEND_OUTPUT_DIR")

reject_parent_ref "$FRONTEND_APP_DIR" FRONTEND_APP_DIR
reject_parent_ref "$FRONTEND_OUTPUT_DIR" FRONTEND_OUTPUT_DIR
reject_parent_ref "$FRONTEND_DIST_DIR" FRONTEND_DIST_DIR

case "$OUTPUT_DIR" in
	"$ROOT_DIR"/*) ;;
	*) fail "FRONTEND_OUTPUT_DIR must stay inside the repository: $OUTPUT_DIR" ;;
esac

if [ ! -d "$APP_DIR" ]; then
	[ -n "$FRONTEND_GIT_URL" ] || fail "set FRONTEND_GIT_URL or create FRONTEND_APP_DIR first"
	require_command git
	mkdir -p "$(dirname -- "$APP_DIR")"
	log "cloning frontend: $FRONTEND_GIT_URL -> $FRONTEND_APP_DIR"
	git clone "$FRONTEND_GIT_URL" "$APP_DIR"
fi

[ -f "$APP_DIR/package.json" ] || fail "package.json not found in $FRONTEND_APP_DIR"

if [ -n "$FRONTEND_GIT_URL" ] && [ -d "$APP_DIR/.git" ]; then
	origin_url=$(git -C "$APP_DIR" remote get-url origin 2>/dev/null || true)
	if [ -n "$origin_url" ] && [ "$origin_url" != "$FRONTEND_GIT_URL" ]; then
		fail "existing $FRONTEND_APP_DIR origin is $origin_url, not $FRONTEND_GIT_URL"
	fi
fi

if [ -n "$FRONTEND_GIT_REF" ]; then
	[ -d "$APP_DIR/.git" ] || fail "FRONTEND_GIT_REF requires $FRONTEND_APP_DIR to be a git checkout"
	log "checking out frontend ref: $FRONTEND_GIT_REF"
	git -C "$APP_DIR" fetch --tags origin
	git -C "$APP_DIR" checkout "$FRONTEND_GIT_REF"
fi

if [ -f "$APP_DIR/pnpm-lock.yaml" ]; then
	package_manager=pnpm
elif [ -f "$APP_DIR/yarn.lock" ]; then
	package_manager=yarn
elif [ -f "$APP_DIR/bun.lockb" ] || [ -f "$APP_DIR/bun.lock" ]; then
	package_manager=bun
else
	package_manager=npm
fi

if [ -z "$FRONTEND_INSTALL_COMMAND" ]; then
	require_command "$package_manager"
	case "$package_manager" in
		npm)
			if [ -f "$APP_DIR/package-lock.json" ] || [ -f "$APP_DIR/npm-shrinkwrap.json" ]; then
				FRONTEND_INSTALL_COMMAND="npm ci"
			else
				FRONTEND_INSTALL_COMMAND="npm install"
			fi
			;;
		pnpm) FRONTEND_INSTALL_COMMAND="pnpm install --frozen-lockfile" ;;
		yarn) FRONTEND_INSTALL_COMMAND="yarn install --frozen-lockfile" ;;
		bun) FRONTEND_INSTALL_COMMAND="bun install --frozen-lockfile" ;;
	esac
fi

if [ -z "$FRONTEND_BUILD_COMMAND" ]; then
	require_command "$package_manager"
	FRONTEND_BUILD_COMMAND="$package_manager run build"
fi

if ! bool_enabled "$FRONTEND_SKIP_INSTALL"; then
	log "installing frontend dependencies: $FRONTEND_INSTALL_COMMAND"
	(cd "$APP_DIR" && sh -lc "$FRONTEND_INSTALL_COMMAND")
fi

log "building frontend: $FRONTEND_BUILD_COMMAND"
(cd "$APP_DIR" && sh -lc "$FRONTEND_BUILD_COMMAND")

if [ -n "$FRONTEND_DIST_DIR" ]; then
	case "$FRONTEND_DIST_DIR" in
		/*) DIST_DIR=$FRONTEND_DIST_DIR ;;
		*) DIST_DIR=$APP_DIR/$FRONTEND_DIST_DIR ;;
	esac
	[ -d "$DIST_DIR" ] || fail "built frontend dist not found: $FRONTEND_DIST_DIR"
else
	DIST_DIR=
	for candidate in dist build out .output/public; do
		if [ -f "$APP_DIR/$candidate/index.html" ]; then
			DIST_DIR=$APP_DIR/$candidate
			FRONTEND_DIST_DIR=$candidate
			break
		fi
	done
	[ -n "$DIST_DIR" ] || fail "could not find built frontend index.html; set FRONTEND_DIST_DIR to the build output directory"
fi

if bool_enabled "$FRONTEND_REQUIRES_INDEX" && [ ! -f "$DIST_DIR/index.html" ]; then
	fail "built frontend dist must contain index.html for SPA fallback"
fi

mkdir -p "$OUTPUT_DIR"
find "$OUTPUT_DIR" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
cp -R "$DIST_DIR"/. "$OUTPUT_DIR"/

log "frontend synced from $FRONTEND_DIST_DIR to $FRONTEND_OUTPUT_DIR"
