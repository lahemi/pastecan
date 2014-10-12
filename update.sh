#!/bin/sh

targ="$HOME/.local/share/pastecan"
src=$(pwd)

cp -r "${src}/htmls" "$targ"
cp -r "${src}/fgs" "$targ"
cp -r "${src}/styles" "$targ"
