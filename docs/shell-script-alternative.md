# Shell Script: 진짜 파이프라인 구현의 왕도

## 서론

cli-pipe가 수백 줄의 Go 코드로 구현하려는 모든 것을 Shell Script로 더 간단하고 강력하게 만들 수 있다. 이 문서는 "기똥찬 Shell Script"로 파이프라인을 구현하는 방법을 제시한다.

## 1. Shell Script의 숨겨진 강력함

### 기본 템플릿

```bash
#!/usr/bin/env bash
# pipeline.sh - 이게 전부다

set -euo pipefail  # 엄격한 에러 처리
IFS=$'\n\t'       # 안전한 필드 구분

# 색상 출력
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# 로깅 함수
log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*" >&2; }
info() { echo -e "${BLUE}[INFO]${NC} $*"; }

# 진행률 표시
progress() {
    local current=$1
    local total=$2
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((width * current / total))
    
    printf "\r${BLUE}["
    printf "%${completed}s" | tr ' ' '='
    printf "%$((width - completed))s" | tr ' ' ' '
    printf "] %d%%${NC}" "$percentage"
}

# 실행 시간 측정
time_it() {
    local start=$(date +%s)
    "$@"
    local end=$(date +%s)
    log "⏱️  Execution time: $((end - start)) seconds"
}

# 정리 함수
cleanup() {
    log "🧹 Cleaning up..."
    # 임시 파일 정리
    rm -f /tmp/pipeline_*
    # 백그라운드 프로세스 정리
    jobs -p | xargs -r kill 2>/dev/null
}
trap cleanup EXIT
```

### cli-pipe vs Shell Script 비교

**cli-pipe 방식:**
```yaml
# pipeline.yaml
name: backup-pipeline
steps:
  - name: extract
    run: tar cf - /data
    output: data.tar
  - name: compress
    run: gzip -9
    input: data.tar
  - name: checksum
    run: sha256sum
    input: data.tar
```

**Shell Script 방식:**
```bash
#!/usr/bin/env bash
# backup-pipeline.sh

log "🚀 Starting backup pipeline"

# 진짜 스트리밍 파이프라인
tar cf - /data | tee >(sha256sum > backup.sha256) | gzip -9 > backup.tar.gz

log "✅ Backup completed"
log "📦 Archive: backup.tar.gz"
log "🔐 Checksum: backup.sha256"
```

## 2. 스트리밍 파이프라인의 진짜 구현

### 명명된 파이프를 이용한 분기

```bash
#!/usr/bin/env bash
# streaming-pipeline.sh

# 명명된 파이프 생성
mkfifo /tmp/pipe_compress /tmp/pipe_checksum /tmp/pipe_analyze

# 백그라운드 소비자들
{
    log "🗜️  Starting compression..."
    gzip -9 < /tmp/pipe_compress > backup.tar.gz
    log "✅ Compression completed"
} &

{
    log "🔐 Calculating checksum..."
    sha256sum < /tmp/pipe_checksum > backup.sha256
    log "✅ Checksum completed"
} &

{
    log "📊 Analyzing data..."
    wc -l < /tmp/pipe_analyze > analysis.txt
    log "✅ Analysis completed"
} &

# 생산자가 모든 소비자에게 데이터 전송
log "📤 Extracting data..."
tar cf - /data | tee /tmp/pipe_compress /tmp/pipe_checksum > /tmp/pipe_analyze

# 모든 작업 완료 대기
wait
log "🎉 All pipeline tasks completed!"

# 정리
rm -f /tmp/pipe_*
```

### 동적 파이프라인 구성

```bash
#!/usr/bin/env bash
# dynamic-pipeline.sh

# 파이프라인 단계 정의
declare -a PIPELINE_STEPS=(
    "extract:tar cf - /data"
    "compress:gzip -9"
    "encrypt:openssl enc -aes-256-cbc -pass pass:secret"
    "upload:aws s3 cp - s3://backup-bucket/$(date +%Y%m%d).tar.gz.enc"
)

# 파이프라인 실행
execute_pipeline() {
    local cmd_chain=""
    
    for step in "${PIPELINE_STEPS[@]}"; do
        local name="${step%%:*}"
        local command="${step#*:}"
        
        if [[ -z "$cmd_chain" ]]; then
            cmd_chain="$command"
        else
            cmd_chain="$cmd_chain | $command"
        fi
        
        log "📋 Added step: $name"
    done
    
    log "🚀 Executing pipeline: $cmd_chain"
    eval "$cmd_chain"
}

execute_pipeline
```

## 3. 병렬 실행과 의존성 관리

### 고급 의존성 관리

```bash
#!/usr/bin/env bash
# dependency-pipeline.sh

# 작업 정의
declare -A JOBS=(
    [download_data]="curl -L https://example.com/data.tar.gz -o data.tar.gz"
    [download_config]="curl -L https://example.com/config.json -o config.json"
    [extract]="tar xf data.tar.gz"
    [validate]="python validate.py config.json"
    [process]="python process.py --config config.json"
    [backup]="cp processed_data.json backup/"
    [upload]="aws s3 cp processed_data.json s3://results/"
)

# 의존성 정의
declare -A DEPENDENCIES=(
    [download_data]=""
    [download_config]=""
    [extract]="download_data"
    [validate]="download_config"
    [process]="extract validate"
    [backup]="process"
    [upload]="process"
)

# 작업 상태 추적
declare -A JOB_STATUS=()
declare -A JOB_PIDS=()

# 작업 실행 함수
execute_job() {
    local job_name="$1"
    local job_cmd="${JOBS[$job_name]}"
    
    log "🚀 Starting job: $job_name"
    JOB_STATUS[$job_name]="running"
    
    # 실제 명령어 실행
    if eval "$job_cmd"; then
        JOB_STATUS[$job_name]="completed"
        log "✅ Job completed: $job_name"
    else
        JOB_STATUS[$job_name]="failed"
        error "❌ Job failed: $job_name"
        return 1
    fi
}

# 의존성 확인
check_dependencies() {
    local job_name="$1"
    local deps="${DEPENDENCIES[$job_name]}"
    
    if [[ -z "$deps" ]]; then
        return 0  # 의존성 없음
    fi
    
    for dep in $deps; do
        if [[ "${JOB_STATUS[$dep]}" != "completed" ]]; then
            return 1  # 의존성 미완료
        fi
    done
    
    return 0  # 모든 의존성 완료
}

# 실행 가능한 작업 찾기
find_ready_jobs() {
    for job in "${!JOBS[@]}"; do
        if [[ "${JOB_STATUS[$job]}" == "" ]] && check_dependencies "$job"; then
            echo "$job"
        fi
    done
}

# 메인 실행 루프
main() {
    local max_parallel=${MAX_PARALLEL:-4}
    local running_jobs=0
    
    # 모든 작업 상태 초기화
    for job in "${!JOBS[@]}"; do
        JOB_STATUS[$job]=""
    done
    
    while true; do
        # 완료된 백그라운드 작업 확인
        for job in "${!JOB_PIDS[@]}"; do
            if ! kill -0 "${JOB_PIDS[$job]}" 2>/dev/null; then
                unset JOB_PIDS[$job]
                ((running_jobs--))
            fi
        done
        
        # 실행 가능한 작업 찾기
        if [[ $running_jobs -lt $max_parallel ]]; then
            local ready_jobs=($(find_ready_jobs))
            
            for job in "${ready_jobs[@]}"; do
                if [[ $running_jobs -lt $max_parallel ]]; then
                    execute_job "$job" &
                    JOB_PIDS[$job]=$!
                    ((running_jobs++))
                fi
            done
        fi
        
        # 모든 작업 완료 확인
        local all_done=true
        for job in "${!JOBS[@]}"; do
            if [[ "${JOB_STATUS[$job]}" != "completed" ]]; then
                all_done=false
                break
            fi
        done
        
        if [[ "$all_done" == true ]]; then
            break
        fi
        
        sleep 1
    done
    
    log "🎉 All jobs completed!"
}

main
```

## 4. 에러 처리와 재시도

### 똑똑한 재시도 메커니즘

```bash
#!/usr/bin/env bash
# retry-pipeline.sh

# 지수 백오프 재시도
retry_with_backoff() {
    local max_attempts="${1:-3}"
    local base_delay="${2:-1}"
    local max_delay="${3:-60}"
    shift 3
    
    local attempt=1
    local delay=$base_delay
    
    while (( attempt <= max_attempts )); do
        log "🔄 Attempt $attempt/$max_attempts: $*"
        
        if "$@"; then
            log "✅ Command succeeded: $*"
            return 0
        fi
        
        if (( attempt == max_attempts )); then
            error "❌ Command failed after $max_attempts attempts: $*"
            return 1
        fi
        
        warn "⏳ Retrying in ${delay}s..."
        sleep "$delay"
        
        # 지수 백오프 (최대 delay까지)
        delay=$((delay * 2))
        if (( delay > max_delay )); then
            delay=$max_delay
        fi
        
        ((attempt++))
    done
}

# 조건부 재시도
retry_on_condition() {
    local condition_check="$1"
    local max_attempts="${2:-5}"
    shift 2
    
    local attempt=1
    
    while (( attempt <= max_attempts )); do
        if "$@"; then
            return 0
        fi
        
        # 조건 확인
        if ! eval "$condition_check"; then
            error "❌ Condition not met, giving up: $condition_check"
            return 1
        fi
        
        warn "🔄 Condition met, retrying ($attempt/$max_attempts)..."
        sleep $((attempt * 2))
        ((attempt++))
    done
    
    return 1
}

# 사용 예제
demo_retry() {
    log "🌐 Testing network operations..."
    
    # 네트워크 다운로드 재시도
    retry_with_backoff 5 2 30 \
        curl -f -L https://example.com/large-file.tar.gz -o data.tar.gz
    
    # 데이터베이스 연결 재시도
    retry_on_condition "ping -c 1 db.example.com >/dev/null" 10 \
        psql -h db.example.com -c "SELECT 1"
    
    log "✅ All retry operations completed"
}
```

### 서킷 브레이커 패턴

```bash
#!/usr/bin/env bash
# circuit-breaker.sh

# 서킷 브레이커 상태
declare -A CIRCUIT_STATE=()
declare -A CIRCUIT_FAILURES=()
declare -A CIRCUIT_LAST_FAILURE=()

# 서킷 브레이커 설정
readonly FAILURE_THRESHOLD=5
readonly RECOVERY_TIMEOUT=60

# 서킷 브레이커 실행
circuit_breaker() {
    local circuit_name="$1"
    shift
    
    local state="${CIRCUIT_STATE[$circuit_name]:-closed}"
    local failures="${CIRCUIT_FAILURES[$circuit_name]:-0}"
    local last_failure="${CIRCUIT_LAST_FAILURE[$circuit_name]:-0}"
    local now=$(date +%s)
    
    case "$state" in
        open)
            if (( now - last_failure > RECOVERY_TIMEOUT )); then
                log "🔄 Circuit $circuit_name: attempting recovery"
                CIRCUIT_STATE[$circuit_name]="half-open"
            else
                error "⚡ Circuit $circuit_name: open (failing fast)"
                return 1
            fi
            ;;
        half-open)
            log "🔍 Circuit $circuit_name: testing in half-open state"
            ;;
        closed)
            log "✅ Circuit $circuit_name: closed (normal operation)"
            ;;
    esac
    
    # 명령어 실행
    if "$@"; then
        # 성공시 서킷 복구
        CIRCUIT_STATE[$circuit_name]="closed"
        CIRCUIT_FAILURES[$circuit_name]=0
        log "✅ Circuit $circuit_name: command succeeded"
        return 0
    else
        # 실패시 서킷 상태 업데이트
        ((failures++))
        CIRCUIT_FAILURES[$circuit_name]=$failures
        CIRCUIT_LAST_FAILURE[$circuit_name]=$now
        
        if (( failures >= FAILURE_THRESHOLD )); then
            CIRCUIT_STATE[$circuit_name]="open"
            error "⚡ Circuit $circuit_name: opened due to failures ($failures)"
        fi
        
        error "❌ Circuit $circuit_name: command failed"
        return 1
    fi
}
```

## 5. 설정과 파라미터 관리

### 환경 기반 설정

```bash
#!/usr/bin/env bash
# config-pipeline.sh

# 기본 설정
declare -A DEFAULT_CONFIG=(
    [SOURCE_DIR]="/data"
    [BACKUP_DIR]="/backup"
    [COMPRESSION_LEVEL]="9"
    [PARALLEL_JOBS]="$(nproc)"
    [RETENTION_DAYS]="30"
    [LOG_LEVEL]="INFO"
    [DRY_RUN]="false"
)

# 설정 파일 로드
load_config() {
    local config_file="${1:-pipeline.conf}"
    
    if [[ -f "$config_file" ]]; then
        log "📋 Loading config from: $config_file"
        # shellcheck source=/dev/null
        source "$config_file"
    fi
}

# 환경 변수 적용
apply_env_config() {
    for key in "${!DEFAULT_CONFIG[@]}"; do
        if [[ -n "${!key:-}" ]]; then
            log "🔧 Using environment variable: $key=${!key}"
        else
            declare -g "$key"="${DEFAULT_CONFIG[$key]}"
            log "📌 Using default value: $key=${DEFAULT_CONFIG[$key]}"
        fi
    done
}

# 명령줄 인자 파싱
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -s|--source)
                SOURCE_DIR="$2"
                shift 2
                ;;
            -d|--dest)
                BACKUP_DIR="$2"
                shift 2
                ;;
            -j|--jobs)
                PARALLEL_JOBS="$2"
                shift 2
                ;;
            -c|--compression)
                COMPRESSION_LEVEL="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            -v|--verbose)
                LOG_LEVEL="DEBUG"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 설정 검증
validate_config() {
    local errors=0
    
    # 디렉토리 검증
    if [[ ! -d "$SOURCE_DIR" ]]; then
        error "Source directory does not exist: $SOURCE_DIR"
        ((errors++))
    fi
    
    # 숫자 검증
    if ! [[ "$COMPRESSION_LEVEL" =~ ^[1-9]$ ]]; then
        error "Invalid compression level: $COMPRESSION_LEVEL (must be 1-9)"
        ((errors++))
    fi
    
    if ! [[ "$PARALLEL_JOBS" =~ ^[1-9][0-9]*$ ]]; then
        error "Invalid parallel jobs: $PARALLEL_JOBS (must be positive integer)"
        ((errors++))
    fi
    
    return $errors
}

# 도움말 표시
show_help() {
    cat << 'EOF'
Usage: pipeline.sh [OPTIONS]

Options:
  -s, --source DIR      Source directory (default: /data)
  -d, --dest DIR        Backup directory (default: /backup)
  -j, --jobs N          Parallel jobs (default: CPU cores)
  -c, --compression N   Compression level 1-9 (default: 9)
      --dry-run         Show what would be done
  -v, --verbose         Enable verbose logging
  -h, --help            Show this help

Environment variables:
  SOURCE_DIR, BACKUP_DIR, COMPRESSION_LEVEL, PARALLEL_JOBS,
  RETENTION_DAYS, LOG_LEVEL, DRY_RUN

Configuration file:
  pipeline.conf (key=value format)
EOF
}

# 메인 설정 초기화
init_config() {
    load_config "$@"
    parse_args "$@"
    apply_env_config
    validate_config
}
```

## 6. 모니터링과 로깅

### 구조화된 로깅

```bash
#!/usr/bin/env bash
# logging-pipeline.sh

# 로그 레벨 정의
declare -A LOG_LEVELS=(
    [DEBUG]=0
    [INFO]=1
    [WARN]=2
    [ERROR]=3
)

# 현재 로그 레벨
CURRENT_LOG_LEVEL=${LOG_LEVEL:-INFO}

# 로그 함수
log_with_level() {
    local level="$1"
    local message="$2"
    shift 2
    
    local level_num=${LOG_LEVELS[$level]}
    local current_num=${LOG_LEVELS[$CURRENT_LOG_LEVEL]}
    
    if (( level_num >= current_num )); then
        local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
        local color
        
        case "$level" in
            DEBUG) color=$'\033[0;37m' ;;  # White
            INFO)  color=$'\033[0;32m' ;;  # Green
            WARN)  color=$'\033[1;33m' ;;  # Yellow
            ERROR) color=$'\033[0;31m' ;;  # Red
        esac
        
        printf "%s[%s] %s%s %s\n" \
            "$color" "$timestamp" "$level" "$NC" "$message" "$@"
    fi
}

debug() { log_with_level DEBUG "$@"; }
info() { log_with_level INFO "$@"; }
warn() { log_with_level WARN "$@"; }
error() { log_with_level ERROR "$@"; }

# JSON 로깅
log_json() {
    local level="$1"
    local message="$2"
    local metadata="${3:-{}}"
    
    jq -n \
        --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")" \
        --arg level "$level" \
        --arg message "$message" \
        --argjson metadata "$metadata" \
        --arg hostname "$(hostname)" \
        --arg pid "$$" \
        '{
            timestamp: $timestamp,
            level: $level,
            message: $message,
            metadata: $metadata,
            hostname: $hostname,
            pid: $pid
        }'
}

# 메트릭 수집
declare -A METRICS=()

metric() {
    local name="$1"
    local value="$2"
    local timestamp=$(date +%s)
    
    METRICS["$name"]="$value"
    
    # 메트릭 로깅
    log_json "METRIC" "Recorded metric: $name" \
        "$(jq -n --arg name "$name" --arg value "$value" --arg timestamp "$timestamp" \
            '{name: $name, value: $value, timestamp: $timestamp}')"
}

# 성능 모니터링
monitor_performance() {
    local interval=${1:-5}
    local logfile="${2:-/tmp/pipeline_metrics.log}"
    
    while true; do
        local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
        local memory_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
        local disk_usage=$(df /tmp | tail -1 | awk '{print $5}' | cut -d'%' -f1)
        
        {
            echo "timestamp=$(date +%s)"
            echo "cpu_usage=$cpu_usage"
            echo "memory_usage=$memory_usage"
            echo "disk_usage=$disk_usage"
            echo "---"
        } >> "$logfile"
        
        sleep "$interval"
    done &
    
    echo $!  # 모니터링 프로세스 PID 반환
}
```

### 실시간 대시보드

```bash
#!/usr/bin/env bash
# dashboard-pipeline.sh

# 대시보드 표시
show_dashboard() {
    local log_file="$1"
    
    while true; do
        clear
        
        # 헤더
        echo -e "${BLUE}================================${NC}"
        echo -e "${BLUE}    Pipeline Dashboard${NC}"
        echo -e "${BLUE}================================${NC}"
        echo
        
        # 시스템 상태
        echo -e "${GREEN}System Status:${NC}"
        echo "  Time: $(date)"
        echo "  Load: $(uptime | awk -F'load average:' '{print $2}')"
        echo "  Memory: $(free -h | grep Mem | awk '{print $3"/"$2}')"
        echo
        
        # 실행 중인 작업
        echo -e "${GREEN}Running Jobs:${NC}"
        jobs -l | while read -r job; do
            echo "  $job"
        done
        echo
        
        # 최근 로그
        echo -e "${GREEN}Recent Logs:${NC}"
        if [[ -f "$log_file" ]]; then
            tail -10 "$log_file"
        fi
        
        sleep 2
    done
}

# 진행률 표시
show_progress_bar() {
    local current="$1"
    local total="$2"
    local width="${3:-50}"
    local label="${4:-Progress}"
    
    local percentage=$((current * 100 / total))
    local filled=$((width * current / total))
    local empty=$((width - filled))
    
    printf "\r%s: [" "$label"
    printf "%*s" "$filled" | tr ' ' '='
    printf "%*s" "$empty" | tr ' ' '-'
    printf "] %d%% (%d/%d)" "$percentage" "$current" "$total"
}

# 실시간 로그 출력
tail_logs() {
    local log_file="$1"
    local pattern="${2:-.*}"
    
    tail -f "$log_file" | while read -r line; do
        if [[ "$line" =~ $pattern ]]; then
            case "$line" in
                *ERROR*) echo -e "${RED}$line${NC}" ;;
                *WARN*)  echo -e "${YELLOW}$line${NC}" ;;
                *INFO*)  echo -e "${GREEN}$line${NC}" ;;
                *)       echo "$line" ;;
            esac
        fi
    done
}
```

## 7. 실제 고급 파이프라인 예제

### 완전한 데이터 처리 파이프라인

```bash
#!/usr/bin/env bash
# complete-pipeline.sh

# 설정 초기화
source "$(dirname "$0")/lib/config.sh"
source "$(dirname "$0")/lib/logging.sh"
source "$(dirname "$0")/lib/retry.sh"

# 파이프라인 정의
declare -A PIPELINE_STEPS=(
    [download]="download_data"
    [validate]="validate_data"
    [transform]="transform_data"
    [analyze]="analyze_data"
    [report]="generate_report"
    [cleanup]="cleanup_temp_files"
)

declare -A STEP_DEPENDENCIES=(
    [download]=""
    [validate]="download"
    [transform]="validate"
    [analyze]="transform"
    [report]="analyze"
    [cleanup]="report"
)

# 단계별 구현
download_data() {
    info "📥 Downloading data..."
    
    local urls=(
        "https://example.com/data1.csv"
        "https://example.com/data2.json"
        "https://example.com/data3.xml"
    )
    
    local pids=()
    for url in "${urls[@]}"; do
        local filename=$(basename "$url")
        retry_with_backoff 3 2 30 \
            curl -L "$url" -o "data/$filename" &
        pids+=($!)
    done
    
    # 모든 다운로드 완료 대기
    for pid in "${pids[@]}"; do
        wait "$pid" || return 1
    done
    
    info "✅ Download completed"
}

validate_data() {
    info "✅ Validating data..."
    
    local errors=0
    
    # CSV 검증
    if ! head -1 data/data1.csv | grep -q "id,name,value"; then
        error "Invalid CSV header"
        ((errors++))
    fi
    
    # JSON 검증
    if ! jq empty data/data2.json; then
        error "Invalid JSON format"
        ((errors++))
    fi
    
    # XML 검증
    if ! xmllint --noout data/data3.xml; then
        error "Invalid XML format"
        ((errors++))
    fi
    
    if (( errors > 0 )); then
        error "Data validation failed with $errors errors"
        return 1
    fi
    
    info "✅ Data validation passed"
}

transform_data() {
    info "🔄 Transforming data..."
    
    # CSV to JSON
    {
        echo "Converting CSV to JSON..."
        csv2json data/data1.csv > processed/data1.json
    } &
    
    # JSON normalization
    {
        echo "Normalizing JSON..."
        jq '.[] | {id: .id, name: .name, value: (.value | tonumber)}' \
            data/data2.json > processed/data2_normalized.json
    } &
    
    # XML to JSON
    {
        echo "Converting XML to JSON..."
        xmlstarlet sel -t -o '{"items":[' \
            -m "//item" -o '{"id":"' -v "@id" -o '","name":"' -v "name" -o '"},' \
            -o ']}' data/data3.xml | sed 's/,]/]}/' > processed/data3.json
    } &
    
    wait  # 모든 변환 완료 대기
    
    info "✅ Data transformation completed"
}

analyze_data() {
    info "📊 Analyzing data..."
    
    # 데이터 통계
    {
        echo "Calculating statistics..."
        jq -s '
            map(select(type == "array") | .[]) |
            group_by(.name) |
            map({
                name: .[0].name,
                count: length,
                avg_value: (map(.value) | add / length)
            })
        ' processed/*.json > analysis/statistics.json
    } &
    
    # 데이터 품질 검사
    {
        echo "Checking data quality..."
        jq -s '
            map(select(type == "array") | .[]) |
            {
                total_records: length,
                null_values: map(select(.value == null)) | length,
                duplicate_ids: (group_by(.id) | map(select(length > 1)) | length)
            }
        ' processed/*.json > analysis/quality.json
    } &
    
    wait
    
    info "✅ Data analysis completed"
}

generate_report() {
    info "📋 Generating report..."
    
    # HTML 리포트 생성
    cat > reports/report.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Pipeline Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .metric { background: #f0f0f0; padding: 10px; margin: 10px 0; }
        .success { color: green; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>Pipeline Execution Report</h1>
    <div class="metric">
        <h2>Execution Time</h2>
        <p>Started: $(date)</p>
        <p>Duration: ${SECONDS} seconds</p>
    </div>
    
    <div class="metric">
        <h2>Data Statistics</h2>
        <pre>$(cat analysis/statistics.json | jq .)</pre>
    </div>
    
    <div class="metric">
        <h2>Data Quality</h2>
        <pre>$(cat analysis/quality.json | jq .)</pre>
    </div>
</body>
</html>
EOF
    
    # 이메일 알림
    if command -v mail >/dev/null; then
        {
            echo "Pipeline execution completed successfully!"
            echo "Report: file://$(pwd)/reports/report.html"
            echo "Duration: ${SECONDS} seconds"
        } | mail -s "Pipeline Report" admin@example.com
    fi
    
    info "✅ Report generated"
}

cleanup_temp_files() {
    info "🧹 Cleaning up temporary files..."
    
    # 7일 이상된 임시 파일 삭제
    find temp/ -type f -mtime +7 -delete
    
    # 로그 파일 압축
    find logs/ -name "*.log" -mtime +1 -exec gzip {} \;
    
    info "✅ Cleanup completed"
}

# 메인 실행
main() {
    info "🚀 Starting complete pipeline..."
    
    # 디렉토리 생성
    mkdir -p {data,processed,analysis,reports,temp,logs}
    
    # 성능 모니터링 시작
    local monitor_pid
    monitor_pid=$(monitor_performance 5 "logs/performance.log")
    
    # 파이프라인 실행
    for step in download validate transform analyze report cleanup; do
        info "▶️  Executing step: $step"
        
        if time_it "${PIPELINE_STEPS[$step]}"; then
            info "✅ Step completed: $step"
        else
            error "❌ Step failed: $step"
            kill "$monitor_pid" 2>/dev/null
            return 1
        fi
    done
    
    # 모니터링 종료
    kill "$monitor_pid" 2>/dev/null
    
    info "🎉 Pipeline completed successfully!"
    info "📊 Total execution time: ${SECONDS} seconds"
}

# 실행
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
```

## 8. Shell Script 프레임워크

### bashpipe - 파이프라인 프레임워크

```bash
#!/usr/bin/env bash
# bashpipe.sh - Shell Script Pipeline Framework

# 프레임워크 초기화
bashpipe::init() {
    set -euo pipefail
    
    # 전역 변수
    declare -g -A BASHPIPE_STEPS=()
    declare -g -A BASHPIPE_DEPS=()
    declare -g -A BASHPIPE_CONFIG=()
    declare -g BASHPIPE_LOGFILE="/tmp/bashpipe.log"
    
    # 기본 설정
    BASHPIPE_CONFIG[MAX_PARALLEL]=4
    BASHPIPE_CONFIG[LOG_LEVEL]="INFO"
    BASHPIPE_CONFIG[RETRY_COUNT]=3
    
    # 로깅 설정
    exec 1> >(tee -a "$BASHPIPE_LOGFILE")
    exec 2> >(tee -a "$BASHPIPE_LOGFILE" >&2)
}

# 파이프라인 정의
bashpipe::pipeline() {
    local name="$1"
    shift
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --description)
                BASHPIPE_CONFIG[DESCRIPTION]="$2"
                shift 2
                ;;
            --on-error)
                BASHPIPE_CONFIG[ON_ERROR]="$2"
                shift 2
                ;;
            --parallel)
                BASHPIPE_CONFIG[MAX_PARALLEL]="$2"
                shift 2
                ;;
            *)
                echo "Unknown option: $1"
                return 1
                ;;
        esac
    done
    
    BASHPIPE_CONFIG[NAME]="$name"
    echo "📋 Pipeline defined: $name"
}

# 단계 정의
bashpipe::step() {
    local step_name="$1"
    shift
    
    local cmd=""
    local input=""
    local output=""
    local deps=""
    local parallel=1
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --run)
                cmd="$2"
                shift 2
                ;;
            --input)
                input="$2"
                shift 2
                ;;
            --output)
                output="$2"
                shift 2
                ;;
            --depends)
                deps="$2"
                shift 2
                ;;
            --parallel)
                parallel="$2"
                shift 2
                ;;
            *)
                echo "Unknown option: $1"
                return 1
                ;;
        esac
    done
    
    BASHPIPE_STEPS["$step_name"]="$cmd"
    BASHPIPE_DEPS["$step_name"]="$deps"
    
    echo "📝 Step defined: $step_name"
}

# 파이프라인 실행
bashpipe::execute() {
    local pipeline_name="$1"
    
    echo "🚀 Executing pipeline: $pipeline_name"
    echo "📖 Description: ${BASHPIPE_CONFIG[DESCRIPTION]:-No description}"
    
    # 의존성 순서로 실행
    local executed=()
    local total_steps=${#BASHPIPE_STEPS[@]}
    local current_step=0
    
    for step in "${!BASHPIPE_STEPS[@]}"; do
        if bashpipe::_can_execute "$step" "${executed[@]}"; then
            ((current_step++))
            
            echo "▶️  Step $current_step/$total_steps: $step"
            
            if bashpipe::_execute_step "$step"; then
                executed+=("$step")
                echo "✅ Step completed: $step"
            else
                echo "❌ Step failed: $step"
                return 1
            fi
        fi
    done
    
    echo "🎉 Pipeline completed: $pipeline_name"
}

# 단계 실행 가능 여부 확인
bashpipe::_can_execute() {
    local step="$1"
    shift
    local executed=("$@")
    
    local deps="${BASHPIPE_DEPS[$step]}"
    
    if [[ -z "$deps" ]]; then
        return 0
    fi
    
    for dep in $deps; do
        if ! printf '%s\n' "${executed[@]}" | grep -q "^$dep$"; then
            return 1
        fi
    done
    
    return 0
}

# 단계 실행
bashpipe::_execute_step() {
    local step="$1"
    local cmd="${BASHPIPE_STEPS[$step]}"
    
    local start_time=$(date +%s)
    
    if eval "$cmd"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        echo "⏱️  Step '$step' completed in ${duration}s"
        return 0
    else
        echo "❌ Step '$step' failed"
        return 1
    fi
}

# 사용 예제
demo_framework() {
    # 프레임워크 초기화
    bashpipe::init
    
    # 파이프라인 정의
    bashpipe::pipeline "data-processing" \
        --description "Process data files" \
        --on-error "stop" \
        --parallel 2
    
    # 단계 정의
    bashpipe::step "download" \
        --run "curl -L https://example.com/data.csv -o data.csv"
    
    bashpipe::step "validate" \
        --run "python validate.py data.csv" \
        --depends "download"
    
    bashpipe::step "process" \
        --run "python process.py data.csv" \
        --depends "validate"
    
    bashpipe::step "backup" \
        --run "cp processed_data.json backup/" \
        --depends "process"
    
    # 실행
    bashpipe::execute "data-processing"
}
```

## 9. 성능 비교

### Shell Script vs cli-pipe

**시작 시간:**
- Shell Script: 즉시 (bash 로딩 시간만)
- cli-pipe: Go 바이너리 로딩 + YAML 파싱 + 검증

**메모리 사용량:**
- Shell Script: 최소 (bash 프로세스만)
- cli-pipe: Go 런타임 + 추가 메모리 할당

**실행 오버헤드:**
- Shell Script: 없음 (네이티브 Unix)
- cli-pipe: Go → bash 변환 + 추가 프로세스

**디버깅:**
- Shell Script: `bash -x` 한 줄로 모든 것 추적
- cli-pipe: Go 디버거 + 로그 분석

## 10. 결론

### Shell Script가 나은 이유

1. **단순함**: 추가 런타임 없음
2. **투명성**: 모든 것이 보임
3. **성능**: 네이티브 Unix 성능
4. **유연성**: 쉘의 모든 기능 사용 가능
5. **디버깅**: 표준 도구로 완벽 추적
6. **이식성**: 모든 Unix 시스템에서 동작

### cli-pipe가 나은 경우

1. **팀 표준화**: "YAML이 더 읽기 쉬워요"
2. **Go 생태계**: 다른 Go 도구와 통합
3. **타입 안전성**: 컴파일 타임 체크
4. **웹 UI**: 향후 웹 인터페이스 계획

### 진짜 결론

> "복잡한 문제를 단순하게 해결하는 것이 천재다." - 알베르트 아인슈타인

cli-pipe가 수백 줄의 Go 코드로 구현하려는 모든 기능을 Shell Script로 더 간단하고 강력하게 만들 수 있다. 

**진짜 Unix 철학:**
- 작은 도구들을 조합하라
- 텍스트 스트림을 사용하라
- 침묵은 금이다
- 실패는 큰 소리로 하라

Shell Script는 이 모든 원칙을 자연스럽게 따른다. cli-pipe는 이를 억지로 재구현한다.

> "The best code is no code at all." - Jeff Atwood

필요 없는 코드를 만들지 말고, 이미 완벽한 도구를 사용하라.

---

*이 문서는 Shell Script의 진짜 강력함을 보여주기 위해 작성되었습니다. 때로는 가장 오래된 도구가 가장 좋은 도구입니다.*