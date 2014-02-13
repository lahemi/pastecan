#!/bin/sh
# Crude installation script, feel free to throw
# things where ever you please. Remember to check
# the Go-sources, too, though; some of the things are
# hard-coded at the moment.

targ="$HOME/.local/share/pastecan"
tmps="/tmp/pastecan"
src=$(pwd)

[ ! -d "targ" ] && mkdir -p "$targ"

cp -r "${src}/htmls" "$targ"
cp -r "${src}/fgs" "$targ"
cp -r "${src}/styles" "$targ"
[ ! -d "$tmps" ] && mkdir "$tmps"

if [ -n "$GOPATH" ]; then
    if [ ! -d "${GOPATH}/src/pastecan" ]; then
        mkdir "${GOPATH}/src/pastecan"
    fi
    cp -r *go "${GOPATH}/src/pastecan" &&
        cp -r pbnf "${GOPATH}/src/pastecan/" &&
    go install pastecan
fi

