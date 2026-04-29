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
    WEB_LAUNCHER="web.sh"
    ARCHIVE_EXT="tar.gz"
    ;;
  macos)
    WEB_BIN="dtool"
    WEB_LAUNCHER="web.command"
    ARCHIVE_EXT="tar.gz"
    ;;
  *)
    echo "[ERROR] unsupported target os: ${TARGET_OS}"
    exit 1
    ;;
esac

PACKAGE_NAME="${2:-}"
if [[ -n "${PACKAGE_NAME}" && "${PACKAGE_NAME}" =~ [\\/:*?\"\<\>\|[:space:]] ]]; then
  echo "[ERROR] package name contains illegal characters: ${PACKAGE_NAME}"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/build"
STAGE_DIR="${BUILD_DIR}/release_${TARGET_OS}"
PACKAGE_DIR="${STAGE_DIR}/package"
if [[ -n "${PACKAGE_NAME}" ]]; then
  PACKAGE_DIR="${STAGE_DIR}/package_${PACKAGE_NAME}"
fi
ARCHIVE_FILE="${BUILD_DIR}/dtool_release_${TARGET_OS}.${ARCHIVE_EXT}"
HOST_UNAME="$(uname -s)"
BUILD_GOOS=""
BUILD_GOARCH=""
BUILD_CGO_ENABLED=""

if [[ "${TARGET_OS}" == "linux" ]]; then
  BUILD_GOOS="linux"
  BUILD_GOARCH="amd64"
  BUILD_CGO_ENABLED="0"
fi

if [[ "${TARGET_OS}" == "macos" ]] && [[ "${HOST_UNAME}" =~ ^(MINGW|MSYS|CYGWIN) ]]; then
  BUILD_GOOS="darwin"
  BUILD_GOARCH="amd64"
  BUILD_CGO_ENABLED="0"
fi

write_step() {
  printf '%s\n' "$1"
}

copy_if_exists() {
  local source_path="$1"
  local target_path="$2"
  local description="$3"

  if [[ -e "${source_path}" ]]; then
    cp "${source_path}" "${target_path}"
  else
    write_step "[WARN] Skip missing ${description}: ${source_path}"
  fi
}

go_build_target() {
  if [[ -n "${BUILD_GOOS}" ]]; then
    GOOS="${BUILD_GOOS}" GOARCH="${BUILD_GOARCH}" CGO_ENABLED="${BUILD_CGO_ENABLED}" go build "$@"
  else
    go build "$@"
  fi
}

mkdir -p "${PACKAGE_DIR}"

write_step "[1/5] Build frontend web/dist"
pushd "${ROOT_DIR}/web" >/dev/null
if [[ -d node_modules/.cache ]]; then
  rm -rf node_modules/.cache
fi
if [[ -f package-lock.json ]]; then
  if ! npm ci; then
    write_step "[WARN] npm ci failed, clearing cache and retrying once"
    rm -rf node_modules/.cache
    npm cache verify
    npm ci --no-audit --no-fund
  fi
else
  npm install --no-audit --no-fund
fi
npm run prod
popd >/dev/null

write_step "[2/5] Build ${TARGET_OS} web backend"
go_build_target -ldflags "-s -w" -o "${PACKAGE_DIR}/${WEB_BIN}" ./cmd/dtool

write_step "[3/5] Copy runtime assets"
mkdir -p "${PACKAGE_DIR}/config/dtool"
mkdir -p "${PACKAGE_DIR}/web/dist"
cp "${ROOT_DIR}/go.mod" "${PACKAGE_DIR}/go.mod"
cp "${ROOT_DIR}/config/dtool/company.ini" "${PACKAGE_DIR}/config/dtool/config.ini"
copy_if_exists "${ROOT_DIR}/config/dtool/frog.db" "${PACKAGE_DIR}/config/dtool/frog.db" "config/dtool/frog.db"
cp -R "${ROOT_DIR}/web/dist" "${PACKAGE_DIR}/web"
mkdir -p "${PACKAGE_DIR}/internal/pkg" "${PACKAGE_DIR}/internal/app/dtool"
cp -R "${ROOT_DIR}/internal/pkg/p_js" "${PACKAGE_DIR}/internal/pkg/p_js"
cp -R "${ROOT_DIR}/internal/app/dtool/database" "${PACKAGE_DIR}/internal/app/dtool/database"
cp -R "${ROOT_DIR}/internal/app/dtool/database_log" "${PACKAGE_DIR}/internal/app/dtool/database_log"

write_step "[4/5] Generate launch scripts and readme"
cat > "${PACKAGE_DIR}/${WEB_LAUNCHER}" <<'EOF'
#!/usr/bin/env bash
set -u

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 默认配置名 / Default config profile name
DEFAULT_CONFIG_NAME="config"
# 运行与日志目录 / Runtime and log directories
LOG_DIR="${SCRIPT_DIR}/logs"
RUN_DIR="${SCRIPT_DIR}/run"
RUNTIME_LOG="${LOG_DIR}/runtime.log"
CRASH_LOG="${LOG_DIR}/crash.log"
SUPERVISOR_PID_FILE="${RUN_DIR}/web.pid"
DTOOL_PID_FILE="${RUN_DIR}/dtool.pid"

# 退避重启参数 / Backoff restart settings
BACKOFF_STEPS=(2 5 10 30)
RESET_WINDOW_SECONDS=60

mkdir -p "${LOG_DIR}" "${RUN_DIR}"

write_runtime_log() {
  local message="$1"
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "${message}" >> "${RUNTIME_LOG}"
}

write_crash_log() {
  local message="$1"
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "${message}" >> "${CRASH_LOG}"
}

is_pid_running() {
  local pid="$1"
  if [[ -z "${pid}" ]]; then
    return 1
  fi
  kill -0 "${pid}" >/dev/null 2>&1
}

read_pid_file() {
  local pid_file="$1"
  if [[ -f "${pid_file}" ]]; then
    tr -d '[:space:]' < "${pid_file}"
  fi
}

remove_pid_file_if_stale() {
  local pid_file="$1"
  local pid=""
  pid="$(read_pid_file "${pid_file}")"
  if [[ -n "${pid}" ]] && ! is_pid_running "${pid}"; then
    rm -f "${pid_file}"
  fi
}

resolve_config_name() {
  local raw_name="${1:-${DEFAULT_CONFIG_NAME}}"
  if [[ "${raw_name}" == *.ini ]]; then
    raw_name="${raw_name%.ini}"
  fi
  printf '%s\n' "${raw_name}"
}

ensure_config_exists() {
  local config_name="$1"
  local config_path="${SCRIPT_DIR}/config/dtool/${config_name}.ini"

  # 启动前必须校验配置存在 / Validate config exists before startup
  if [[ ! -f "${config_path}" ]]; then
    echo "[ERROR] config file not found: ${config_path}" >&2
    exit 1
  fi
}

open_browser() {
  sleep 2
  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "http://localhost:17170/" >/dev/null 2>&1 || true
  elif command -v open >/dev/null 2>&1; then
    open "http://localhost:17170/" >/dev/null 2>&1 || true
  fi
}

stop_process_by_pid_file() {
  local pid_file="$1"
  local process_name="$2"
  local pid=""

  pid="$(read_pid_file "${pid_file}")"
  if [[ -z "${pid}" ]]; then
    return 0
  fi

  if ! is_pid_running "${pid}"; then
    rm -f "${pid_file}"
    return 0
  fi

  kill "${pid}" >/dev/null 2>&1 || true
  for _ in 1 2 3 4 5; do
    if ! is_pid_running "${pid}"; then
      rm -f "${pid_file}"
      return 0
    fi
    sleep 1
  done

  write_crash_log "${process_name} pid=${pid} did not exit in time, sending SIGKILL."
  kill -9 "${pid}" >/dev/null 2>&1 || true
  rm -f "${pid_file}"
}

run_supervisor() {
  local config_name="$1"
  local dtool_pid=""
  local crash_count=0
  local backoff_index=0
  local start_ts=0
  local end_ts=0
  local run_duration=0
  local wait_seconds=0

  trap 'write_crash_log "Supervisor received stop signal."; stop_process_by_pid_file "${DTOOL_PID_FILE}" "dtool"; rm -f "${SUPERVISOR_PID_FILE}"; exit 0' INT TERM

  echo "$$" > "${SUPERVISOR_PID_FILE}"
  write_crash_log "Supervisor started with config=${config_name}."

  while true; do
    start_ts="$(date +%s)"
    write_runtime_log "Starting dtool with config=${config_name}."
    "${SCRIPT_DIR}/dtool" --ConfigFile="${config_name}" >> "${RUNTIME_LOG}" 2>> "${CRASH_LOG}" &
    dtool_pid=$!
    echo "${dtool_pid}" > "${DTOOL_PID_FILE}"
    wait "${dtool_pid}"
    exit_code=$?
    end_ts="$(date +%s)"
    run_duration=$((end_ts - start_ts))
    rm -f "${DTOOL_PID_FILE}"

    # 正常退出时停止守护 / Stop supervising after clean exit
    if [[ ${exit_code} -eq 0 ]]; then
      write_crash_log "dtool exited normally with code=0, supervisor will stop."
      rm -f "${SUPERVISOR_PID_FILE}"
      exit 0
    fi

    # 长时间稳定运行后重置退避 / Reset backoff after stable runtime
    if [[ ${run_duration} -ge ${RESET_WINDOW_SECONDS} ]]; then
      crash_count=0
      backoff_index=0
    fi

    crash_count=$((crash_count + 1))
    backoff_index=$((crash_count - 1))
    if [[ ${backoff_index} -gt $((${#BACKOFF_STEPS[@]} - 1)) ]]; then
      backoff_index=$((${#BACKOFF_STEPS[@]} - 1))
    fi
    wait_seconds="${BACKOFF_STEPS[${backoff_index}]}"

    write_crash_log "dtool crashed with code=${exit_code}, runDuration=${run_duration}s, restartCount=${crash_count}, restartAfter=${wait_seconds}s."
    sleep "${wait_seconds}"
  done
}

start_service() {
  local config_name="$1"
  local supervisor_pid=""

  ensure_config_exists "${config_name}"
  remove_pid_file_if_stale "${SUPERVISOR_PID_FILE}"
  remove_pid_file_if_stale "${DTOOL_PID_FILE}"
  supervisor_pid="$(read_pid_file "${SUPERVISOR_PID_FILE}")"

  # 已运行时拒绝重复启动 / Prevent duplicate start while already running
  if [[ -n "${supervisor_pid}" ]] && is_pid_running "${supervisor_pid}"; then
    echo "web is already running, pid=${supervisor_pid}"
    exit 0
  fi

  nohup "$0" __supervise "${config_name}" >/dev/null 2>&1 &
  sleep 1

  supervisor_pid="$(read_pid_file "${SUPERVISOR_PID_FILE}")"
  if [[ -z "${supervisor_pid}" ]] || ! is_pid_running "${supervisor_pid}"; then
    echo "[ERROR] failed to start web supervisor" >&2
    exit 1
  fi

  echo "web started, pid=${supervisor_pid}, config=${config_name}"
  open_browser
}

stop_service() {
  stop_process_by_pid_file "${SUPERVISOR_PID_FILE}" "web supervisor"
  stop_process_by_pid_file "${DTOOL_PID_FILE}" "dtool"
  echo "web stopped"
}

show_usage() {
  cat <<'USAGE'
Usage:
  ./web.sh start [configName|configName.ini]
  ./web.sh stop
USAGE
}

main() {
  local command="${1:-start}"
  local config_name=""

  case "${command}" in
    start)
      config_name="$(resolve_config_name "${2:-${DEFAULT_CONFIG_NAME}}")"
      start_service "${config_name}"
      ;;
    stop)
      stop_service
      ;;
    __supervise)
      config_name="$(resolve_config_name "${2:-${DEFAULT_CONFIG_NAME}}")"
      run_supervisor "${config_name}"
      ;;
    *)
      show_usage
      exit 1
      ;;
  esac
}

main "$@"
EOF

cat > "${PACKAGE_DIR}/README_RELEASE.txt" <<EOF
dtool release package (${TARGET_OS})

Run web mode:
  bash ${WEB_LAUNCHER} start [configName]
  bash ${WEB_LAUNCHER} stop

Notes:
1. ConfigFile matches config/dtool/*.ini filename without extension, for example: bash ${WEB_LAUNCHER} start company
2. Runtime logs are written to logs/runtime.log, crash and restart logs are written to logs/crash.log
3. Check webPath/dbPath and other ini settings before first run
EOF

chmod +x "${PACKAGE_DIR}/${WEB_BIN}" "${PACKAGE_DIR}/${WEB_LAUNCHER}"

write_step "[5/5] Create archive"
rm -f "${ARCHIVE_FILE}"
tar -czf "${ARCHIVE_FILE}" -C "${PACKAGE_DIR}" .

printf '\n[OK] Package created: %s\n' "${ARCHIVE_FILE}"
