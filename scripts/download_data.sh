#!/bin/bash
# Download the arkhamdb-json-data repository to .data/

set -e

DATA_DIR="${1:-.data/arkhamdb-json-data}"
REPO_URL="https://github.com/Kamalisk/arkhamdb-json-data.git"

if [ -d "$DATA_DIR/.git" ]; then
    echo "Repository already exists at $DATA_DIR. Updating..."
    cd "$DATA_DIR"
    git pull
    cd - > /dev/null
else
    echo "Cloning repository to $DATA_DIR..."
    mkdir -p "$(dirname "$DATA_DIR")"
    git clone "$REPO_URL" "$DATA_DIR"
fi

echo "Data downloaded successfully to $DATA_DIR"

