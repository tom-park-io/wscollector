#!/bin/bash

URL="http://wscollectorapi-lb-219729844.ap-northeast-2.elb.amazonaws.com/api/v1/bybit/kline/final-perf?interval=1m&start=1745573568000&pre_hours=4"
# JSON을 API로부터 받아오기
response=$(curl -s $URL) # 여기에 실제 API URL 삽입

# 전체 key 개수 출력
key_count=$(echo "$response" | jq 'keys | length')
echo "총 key 개수: $key_count"

# 각 key 별로 배열을 반복 처리
echo "$response" | jq -c 'to_entries[]' | while read -r entry; do
    key=$(echo "$entry" | jq -r '.key')
    echo "▶ key: $key"

    echo "$entry" | jq -c '.value[]' | while read -r row; do
        val0=$(echo "$row" | jq '.[0]')
        val1=$(echo "$row" | jq '.[1]')
        # echo "  row: $row"
        # echo "    value 0: $val0"
        # echo "    value 1: $val1"
    done
done
