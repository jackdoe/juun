#!/bin/bash
_realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}
ROOT=$(_realpath $(dirname $BASH_SOURCE))

echo source $ROOT/setup.sh >> $HOME/.bash_profile
echo source $ROOT/setup.sh >> $HOME/.zshrc
