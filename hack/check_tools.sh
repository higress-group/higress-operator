#!/bin/bash

# 检查命令行工具
function install_tools() {
    echo "installing $1 ..."
    if [ "$(uname)" == "Darwin" ]; then
      brew install "$1"
    elif [ "$(expr) substr $(uname -s) 1 5" == "Linux" ]; then
      sudo apt-get install "$1"
    fi
}

# 判断命令行工具是否已经安装
function check_tool() {
  if ! command -v "$1" &> /dev/null; then
    install_tools "$1"
  else
    echo "$1 has already been installed"
  fi
}

# kind
check_tool kind