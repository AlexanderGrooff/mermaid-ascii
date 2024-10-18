#!/bin/bash

set -e
FILE=$1

if [ -z "$FILE" ]; then
    echo "Usage: $0 <file>"
    exit 1
fi

echo "Fixing $FILE"

GRAPH=$(cat $FILE | sed -n -e "/graph/,/---/ p" | head -n-1)

if [ -z "$GRAPH" ]; then
    echo "No graph found in $FILE"
    exit 1
fi

echo "$GRAPH" > $FILE
echo "---" >> $FILE
echo "$GRAPH" | go run main.go ${@:2} >> $FILE
