#!/bin/bash
set -e # 에러 발생하면 즉시 종료

# ======================================
# 기본 설정 (고정값)

# 환경 설정: 'local' 또는 'prod' 직접 설정
ENV="local"
DRY_RUN=false
REPEAT_HOURS=-1

# --------------------------------------
# 공통 변수
# --------------------------------------

PORT=5432
DBNAME="wscollector"
TABLENAME="kline_record"
# require, disable
SSLMODE="disable"
TIMEZONE="Asia/Seoul"

# --------------------------------------
# 로컬 환경용 변수 (ENV=local일 때 사용)
# --------------------------------------

LOCAL_DB_HOST="localhost"
LOCAL_DB_USER="postgres"
LOCAL_DB_PASSWORD="postgres"

# --------------------------------------
# 바이낸스 변수
# --------------------------------------

CATEGORY="linear"
LIMIT=1000
BYBIT_API_URL="https://api.bybit.com/v5/market/kline"
BYBIT_SYMBOL_URL="https://api.bybit.com/v5/market/instruments-info?category=$CATEGORY&limit=$LIMIT"
INTERVAL="1" # 1분

# 현재 시간 (밀리초 단위)
CURRENT_TIME_MS=$(($(date +%s) * 1000))

# 시간 상수
HOUR_MS=3600000
EXTRA_MS=600000
# 안정성을 위해 1시간 보다 조금 더 길게 범위를 잡음
SHIFT_MS=$((HOUR_MS + EXTRA_MS))

# ======================================

# ======================================
# --------------------------------------
# 공통 함수
# --------------------------------------

sleep_seconds() {
    local seconds="$1"
    echo "⏳ ${seconds}초 대기 중..."
    sleep "$seconds"
}
# 타임스탬프 변환 (GNU/BSD 호환) TZ : Asia/Seoul
format_unix_ms_precise() {
    local ms="$1"
    local sec=$((ms / 1000))
    local ms_rem=$((ms % 1000))

    if date --version >/dev/null 2>&1; then
        # GNU date
        base=$(TZ="Asia/Seoul" date -d "@$sec" "+%Y-%m-%d %H:%M:%S")
    else
        # BSD/macOS date
        base=$(TZ="Asia/Seoul" date -r "$sec" "+%Y-%m-%d %H:%M:%S")
    fi

    printf "%s.%03d" "$base" "$ms_rem"
}
now_ms() {
    if command -v gdate >/dev/null 2>&1; then
        # Mac + gdate 설치된 경우
        gdate +%s%3N
    else
        # Linux (GNU date)
        date +%s%3N
    fi
}

# aws(ssm)
get_parameter() {
    local PARAM_NAME="$1"
    local DECRYPT="$2"

    VALUE=$(timeout 5 aws ssm get-parameter --name "$PARAM_NAME" --with-decryption "$DECRYPT" | jq -r '.Parameter.Value')

    if [ "$VALUE" == "null" ] || [ -z "$VALUE" ]; then
        echo ""
    else
        echo "$VALUE"
    fi
}

get_value() {
    local KEY="$1"
    local DEFAULT="$2"

    if [ "$ENV" = "local" ]; then
        case "$KEY" in
        "DJANGO_DEFAULT_DB_HOST")
            echo "$LOCAL_DB_HOST"
            ;;
        "DJANGO_DEFAULT_DB_USER")
            echo "$LOCAL_DB_USER"
            ;;
        "DJANGO_DEFAULT_DB_PASSWORD")
            echo "$LOCAL_DB_PASSWORD"
            ;;
        *)
            echo "$DEFAULT"
            ;;
        esac
    else
        get_parameter "$KEY" true
    fi
}

execute_sql() {
    local SQL="$1"
    local SINGLE_RESULT_ONLY="${2:-false}"

    result=$(echo "$SQL" | PGPASSWORD="$DB_PASSWORD" psql \
        --host="$DB_HOST" \
        --port="$PORT" \
        --username="$DB_USER" \
        --dbname="$DBNAME" \
        --no-align \
        --tuples-only \
        --quiet \
        -v ON_ERROR_STOP=1 2>&1)

    if [[ $? -ne 0 ]]; then
        echo "❌ [ERROR] SQL 실행 실패"
        echo "🚨 실패한 SQL:"
        echo "$SQL"
        echo "$result"
        sleep_seconds 5
        return 1
    fi

    if [ "$SINGLE_RESULT_ONLY" = "true" ]; then
        echo "$result" | head -n 1 | tr -d '[:space:]'
    else
        echo "$result"
    fi
}

# ======================================

# ======================================
# --------------------------------------
# 메인 로직
# --------------------------------------

setup_env() {
    if [ "$ENV" = "prod" ]; then
        echo "🚀 prod 환경 설정 중..."

        # AWS SSM Parameter Store
    else
        echo "🚀 local 환경 설정 중..."

        connect_db
    fi
}

connect_db() {
    # DB 연결
    echo "🔎 DB Connect..."

    # DB 접속 정보 가져오기
    DB_HOST=$(get_value "DJANGO_DEFAULT_DB_HOST" "$LOCAL_DB_HOST")
    DB_USER=$(get_value "DJANGO_DEFAULT_DB_USER" "$LOCAL_DB_USER")
    DB_PASSWORD=$(get_value "DJANGO_DEFAULT_DB_PASSWORD" "$LOCAL_DB_PASSWORD")

    # 필수 값 체크
    if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ]; then
        echo "[Error] DB 접속 정보를 가져오는데 실패했습니다."
        exit 1
    fi

    echo "✅ DB 접속 정보 설정 완료"

    # DSN 문자열은 참고용 (지금은 psql 직접 사용)
    DSN="host=${DB_HOST} port=${PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DBNAME} sslmode=${SSLMODE}"
    if [ -n "$TIMEZONE" ]; then
        DSN+=" TimeZone=${TIMEZONE}"
    fi
    echo "DSN: $DSN"
}

ping_db() {
    # DB Ping 체크
    echo "🔎 DB Ping 테스트 중..."

    PING_RESULT=$(execute_sql "SELECT 1;" true)
    echo "[DEBUG] PING_RESULT='$PING_RESULT'"
    if [ "$PING_RESULT" = "1" ]; then
        echo "✅ DB 연결 성공"
    else
        echo "❌ DB 연결 실패"
        exit 1
    fi
}

task_symbol_api() {
    # Bybit 심볼 리스트 가져오기
    echo "🔎 심볼 리스트 가져오는 중..."

    # 심볼 데이터 전체 가져오기
    RAW_SYMBOLS=$(curl -s "$BYBIT_SYMBOL_URL")

    # JSON 구조 제대로 왔는지 체크
    if ! echo "$RAW_SYMBOLS" | jq -e '.result.list' >/dev/null; then
        echo "❌ 심볼 리스트 가져오기 실패"
        exit 1
    fi

    SYMBOLS=()
    SEEN_BASECOINS=()

    is_seen() {
        local coin="$1"
        for base in "${SEEN_BASECOINS[@]}"; do
            if [ "$base" = "$coin" ]; then
                return 0
            fi
        done
        return 1
    }

    # JSON 한줄씩 읽기 (subshell 문제 없음)
    while IFS= read -r symbol_json; do
        read -r QUOTE_COIN BASE_COIN SYMBOL_NAME <<<"$(echo "$symbol_json" | jq -r '[.quoteCoin, .baseCoin, .symbol] | @tsv')"

        if [ "$QUOTE_COIN" = "USDT" ]; then
            if ! is_seen "$BASE_COIN"; then
                SYMBOLS+=("$SYMBOL_NAME")
                SEEN_BASECOINS+=("$BASE_COIN")
            fi
        fi
    done <<<"$(echo "$RAW_SYMBOLS" | jq -c '.result.list[]')"

    # 🔥 배열 개수
    echo "✅ 최종 SYMBOL 개수: ${#SYMBOLS[@]}"
    # # echo $SYMBOLS
    # for s in "${SYMBOLS[@]}"; do
    #     echo "Symbol: $s"
    # done
}

task_kline_api() {
    # KLINE 수집
    echo "📝 실제 INSERT 실행 중"

    if [ -z "$INPUT_TIME_MS" ]; then
        END_TIME_MS="$CURRENT_TIME_MS"
    else
        END_TIME_MS="$INPUT_TIME_MS"
    fi

    REPEAT_INDEX=0
    while true; do

        # 반복 제한 (유한 반복이면 제한 걸기)
        if [ "$REPEAT_HOURS" -ne -1 ] && [ "$REPEAT_INDEX" -ge "$REPEAT_HOURS" ]; then
            break
        fi

        START_TIME_MS=$((END_TIME_MS - SHIFT_MS))
        # SYMBOLS 배열 반복
        for SYMBOL in "${SYMBOLS[@]}"; do
            echo "🚀 심볼 처리: $SYMBOL"

            echo "⏳ $((REPEAT_INDEX + 1))회차, 요청 중: $SYMBOL | From $(format_unix_ms_precise $START_TIME_MS) To $(format_unix_ms_precise $END_TIME_MS)"

            # Bybit API 호출
            RESPONSE=$(curl -s --max-time 10 "${BYBIT_API_URL}?category=${CATEGORY}&symbol=${SYMBOL}&interval=${INTERVAL}&start=${START_TIME_MS}&end=${END_TIME_MS}&limit=${LIMIT}")

            if echo "$RESPONSE" | jq -e '.result.list' >/dev/null; then
                DATA_COUNT=$(jq '.result.list | length' <<<"$RESPONSE")

                if [ "$DATA_COUNT" -eq 0 ]; then
                    echo "⚠️  데이터 없음 → $SYMBOL 심볼 종료"
                    break
                fi

                echo "✅ 데이터 수신: ${DATA_COUNT}개"

                # 🔥 여기부터 INSERT 작업 시작 (1SYMBOL, 1HOUR)
                jq -c '.result.list[]' <<<"$RESPONSE" | while IFS= read -r row; do
                    start_ms=$(echo "$row" | jq -r '.[0]')
                    open=$(echo "$row" | jq -r '.[1]')
                    high=$(echo "$row" | jq -r '.[2]')
                    low=$(echo "$row" | jq -r '.[3]')
                    close=$(echo "$row" | jq -r '.[4]')
                    volume=$(echo "$row" | jq -r '.[5]')
                    turnover=$(echo "$row" | jq -r '.[6]')

                    # 유효성 검증
                    if [ -z "$open" ] || [ -z "$high" ] || [ -z "$low" ] || [ -z "$close" ] || [ -z "$volume" ] || [ -z "$turnover" ]; then
                        echo "❌ 데이터 누락: $SYMBOL $start_ms"
                        continue
                    fi

                    if [ "$start_ms" -le 0 ]; then
                        echo "❌ 이상한 시작시간: $SYMBOL $start_ms"
                        continue
                    fi

                    start_time=$(format_unix_ms_precise "$start_ms")
                    end_ms=$((start_ms + (INTERVAL * 60 * 1000) - 1))
                    end_time=$(format_unix_ms_precise "$end_ms")
                    now_ms=$(now_ms)
                    timestamp_time=$(format_unix_ms_precise "$now_ms")

                    SQL="INSERT INTO ${TABLENAME} (symbol, interval, start, confirm, \"end\", open, close, high, low, volume, turnover, timestamp, recorded_at)
                VALUES
                ('$SYMBOL', '${INTERVAL}m', '$start_time', true, '$end_time', $open, $close, $high, $low, $volume, $turnover, '$timestamp_time', '$timestamp_time')
                ON CONFLICT (symbol, interval, start, confirm) DO NOTHING
                RETURNING symbol;"

                    EXEC_RESULT=$(execute_sql "$SQL" true)
                    if [ -z "$EXEC_RESULT" ]; then
                        echo "⚠️ CONFLICT 발생 — INSERT는 무시됨 $SYMBOL ${INTERVAL}m $start_time"
                    else
                        echo "✅ INSERT 성공: $EXEC_RESULT ${INTERVAL}m $start_time"
                    fi
                done
                # 🔥 INSERT 끝

            else
                echo "❌ 요청 실패 또는 JSON 파싱 실패 → $SYMBOL 심볼 종료"
                break
            fi
        done

        # 다음 루프 준비
        END_TIME_MS=$((END_TIME_MS - HOUR_MS))
        REPEAT_INDEX=$((REPEAT_INDEX + 1))

    done

    echo "🎉 KLINE 수집 작업 완료!"
}

task_kline_api_dry() {
    # KLINE 수집 테스트
    echo "🔍 [DRY-RUN] INSERT 시뮬레이션 중"

    if [ -z "$INPUT_TIME_MS" ]; then
        END_TIME_MS="$CURRENT_TIME_MS"
    else
        END_TIME_MS="$INPUT_TIME_MS"
    fi

    REPEAT_INDEX=0
    while true; do

        # 반복 제한 (유한 반복이면 제한 걸기)
        if [ "$REPEAT_HOURS" -ne -1 ] && [ "$REPEAT_INDEX" -ge "$REPEAT_HOURS" ]; then
            break
        fi

        START_TIME_MS=$((END_TIME_MS - SHIFT_MS))
        # SYMBOLS 배열 반복
        for SYMBOL in "${SYMBOLS[@]}"; do
            echo "🚀 심볼 처리: $SYMBOL"

            echo "⏳ $((REPEAT_INDEX + 1))회차, 요청 중: $SYMBOL | From $(format_unix_ms_precise $START_TIME_MS) To $(format_unix_ms_precise $END_TIME_MS)"

            # Bybit API 호출
            RESPONSE=$(curl -s --max-time 10 "${BYBIT_API_URL}?category=${CATEGORY}&symbol=${SYMBOL}&interval=${INTERVAL}&start=${START_TIME_MS}&end=${END_TIME_MS}&limit=${LIMIT}")

            if echo "$RESPONSE" | jq -e '.result.list' >/dev/null; then
                DATA_COUNT=$(jq '.result.list | length' <<<"$RESPONSE")

                if [ "$DATA_COUNT" -eq 0 ]; then
                    echo "⚠️  데이터 없음 → $SYMBOL 심볼 종료"
                    break
                fi

                echo "✅ 데이터 수신: ${DATA_COUNT}개"

                # 🔥 여기부터 INSERT 작업 시작 (1SYMBOL, 1HOUR)
                jq -c '.result.list[]' <<<"$RESPONSE" | while IFS= read -r row; do
                    start_ms=$(echo "$row" | jq -r '.[0]')
                    open=$(echo "$row" | jq -r '.[1]')
                    high=$(echo "$row" | jq -r '.[2]')
                    low=$(echo "$row" | jq -r '.[3]')
                    close=$(echo "$row" | jq -r '.[4]')
                    volume=$(echo "$row" | jq -r '.[5]')
                    turnover=$(echo "$row" | jq -r '.[6]')

                    # 유효성 검증
                    if [ -z "$open" ] || [ -z "$high" ] || [ -z "$low" ] || [ -z "$close" ] || [ -z "$volume" ] || [ -z "$turnover" ]; then
                        echo "❌ 데이터 누락: $SYMBOL $start_ms"
                        continue
                    fi

                    if [ "$start_ms" -le 0 ]; then
                        echo "❌ 이상한 시작시간: $SYMBOL $start_ms"
                        continue
                    fi

                    start_time=$(format_unix_ms_precise "$start_ms")
                    end_ms=$((start_ms + (INTERVAL * 60 * 1000) - 1))
                    end_time=$(format_unix_ms_precise "$end_ms")
                    now_ms=$(now_ms)
                    timestamp_time=$(format_unix_ms_precise "$now_ms")

                    echo "✅ INSERT(DRY-RUN, 실제 INSERT X) 성공: $EXEC_RESULT ${INTERVAL}m $start_time"
                done
                # 🔥 INSERT 끝

            else
                echo "❌ 요청 실패 또는 JSON 파싱 실패 → $SYMBOL 심볼 종료"
                break
            fi
        done

        # 다음 루프 준비
        END_TIME_MS=$((END_TIME_MS - HOUR_MS))
        REPEAT_INDEX=$((REPEAT_INDEX + 1))

    done

    echo "🎉 KLINE 수집 작업 완료!"
}

parse_flags() {
    while [[ "$#" -gt 0 ]]; do
        case "$1" in
        -e | --env)
            ENV="$2"
            shift 2
            ;;
        -t | --time)
            INPUT_TIME_MS="$2"
            shift 2
            ;;
        -r | --repeat)
            REPEAT_HOURS="$2"
            shift 2
            ;;
        -d | --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h | --help)
            echo "🛠 사용법:"
            echo "  ./collect_kline.sh [옵션]"
            echo ""
            echo "옵션:"
            echo "  -e, --env    실행 환경 선택 (local, prod)"
            echo "  -t, --time   시작 시간(ms) 입력 (기본: 현재 시간)"
            echo "  -r, --repeat    N시간 반복 실행 (예: 6시간 수집)"
            echo "  -d, --dry-run 실제 실행 없이 시뮬레이션만 수행"
            echo "  -h, --help   도움말 출력"
            exit 0
            ;;
        *)
            echo "❌ 알 수 없는 옵션: $1"
            exit 1
            ;;
        esac
    done
}

main() {
    parse_flags "$@"

    setup_env
    ping_db
    echo "✅ 모든 준비 완료! 작업 시작"

    task_symbol_api
    if [ "$DRY_RUN" = true ]; then
        task_kline_api_dry
    else
        task_kline_api
    fi
    echo "✅ 모든 작업 완료! 스크립트 종료"
}

# === 스크립트 실행 ===
main "$@"
