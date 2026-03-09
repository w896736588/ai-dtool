#!/usr/bin/env bash

set -euo pipefail

TARGET_OS="${1:-}"
if [[ -z "${TARGET_OS}" ]]; then
  echo "[ERROR] missing target os: linux or macos"
  exit 1
fi

case "${TARGET_OS}" in
  linux)
    WEB_BIN="dtool"
    DESKTOP_BIN="dtool_wails"
    WEB_LAUNCHER="web.sh"
    DESKTOP_LAUNCHER="desktop.sh"
    ARCHIVE_EXT="tar.gz"
    ;;
  macos)
    WEB_BIN="dtool"
    DESKTOP_BIN="dtool_wails"
    WEB_LAUNCHER="web.command"
    DESKTOP_LAUNCHER="desktop.command"
    ARCHIVE_EXT="tar.gz"
    ;;
  *)
    echo "[ERROR] unsupported target os: ${TARGET_OS}"
    exit 1
    ;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/build"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
STAGE_DIR="${BUILD_DIR}/release_${TARGET_OS}_${TIMESTAMP}"
PACKAGE_DIR="${STAGE_DIR}/package"
ARCHIVE_FILE="${BUILD_DIR}/dtool_release_${TARGET_OS}_${TIMESTAMP}.${ARCHIVE_EXT}"

write_step() {
  printf '%s\n' "$1"
}

mkdir -p "${PACKAGE_DIR}"

write_step "[1/6] 构建前端 web/dist"
pushd "${ROOT_DIR}/web" >/dev/null
if [[ -d node_modules/.cache ]]; then
  rm -rf node_modules/.cache
fi
if [[ -f package-lock.json ]]; then
  if ! npm ci; then
    write_step "[WARN] npm ci 失败，清理缓存后重试一次"
    rm -rf node_modules/.cache
    npm cache verify
    npm ci --no-audit --no-fund
  fi
else
  npm install --no-audit --no-fund
fi
npm run prod
popd >/dev/null

write_step "[2/6] 构建 ${TARGET_OS} Web 模式后端"
go build -ldflags "-s -w" -o "${PACKAGE_DIR}/${WEB_BIN}" ./cmd/dtool

write_step "[3/6] 构建 ${TARGET_OS} 桌面端"
go build -tags production -ldflags "-s -w" -o "${PACKAGE_DIR}/${DESKTOP_BIN}" ./cmd/dtool_wails

write_step "[4/6] 复制运行资源"
mkdir -p "${PACKAGE_DIR}/config/dtool"
cp "${ROOT_DIR}/go.mod" "${PACKAGE_DIR}/go.mod"
cp "${ROOT_DIR}/config/dtool/company.ini" "${PACKAGE_DIR}/config/dtool/config.ini"
cp "${ROOT_DIR}/config/dtool/frog.db" "${PACKAGE_DIR}/config/dtool/frog.db"
cp -R "${ROOT_DIR}/web/dist" "${PACKAGE_DIR}/web"
mkdir -p "${PACKAGE_DIR}/internal/pkg" "${PACKAGE_DIR}/internal/app/dtool"
cp -R "${ROOT_DIR}/internal/pkg/p_js" "${PACKAGE_DIR}/internal/pkg/p_js"
cp -R "${ROOT_DIR}/internal/app/dtool/database" "${PACKAGE_DIR}/internal/app/dtool/database"

write_step "[5/6] 生成启动脚本和说明文件"
cat > "${PACKAGE_DIR}/${WEB_LAUNCHER}" <<'EOF'
#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
"${SCRIPT_DIR}/dtool" --ConfigFile=config &
sleep 2
if command -v xdg-open >/dev/null 2>&1; then
  xdg-open "http://localhost:17170/" >/dev/null 2>&1 || true
elif command -v open >/dev/null 2>&1; then
  open "http://localhost:17170/" >/dev/null 2>&1 || true
fi
EOF

cat > "${PACKAGE_DIR}/${DESKTOP_LAUNCHER}" <<'EOF'
#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
exec "${SCRIPT_DIR}/dtool_wails" --ConfigFile=config
EOF

cat > "${PACKAGE_DIR}/README_RELEASE.txt" <<EOF
dtool release package (${TARGET_OS})

Run web mode:
  bash ${WEB_LAUNCHER}

Run desktop mode:
  bash ${DESKTOP_LAUNCHER}

Notes:
1. ConfigFile matches config/dtool/*.ini filename without extension
2. Check webPath/dbPath and other ini settings before first run
EOF

chmod +x "${PACKAGE_DIR}/${WEB_BIN}" "${PACKAGE_DIR}/${DESKTOP_BIN}" "${PACKAGE_DIR}/${WEB_LAUNCHER}" "${PACKAGE_DIR}/${DESKTOP_LAUNCHER}"

write_step "[6/6] 压缩归档"
rm -f "${ARCHIVE_FILE}"
tar -czf "${ARCHIVE_FILE}" -C "${PACKAGE_DIR}" .

printf '\n[OK] Package created: %s\n' "${ARCHIVE_FILE}"
