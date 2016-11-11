#!/bin/bash

set -e

TARGETS=(
    "darwin_amd64" 
    "windows_386" 
    "windows_amd64" 
    "linux_386" 
    "linux_amd64"
)

VERSION="$(cat version.txt | tr -d '[[:space:]]')"

for target in "${TARGETS[@]}"; do
    IFS="_"
    set $target
    IFS=" "
    os=$1
    arch=$2
    wd="dist/${os}_${arch}"
    echo Building in ${wd} ...
    mkdir -p ${wd}
    pushd ${wd} > /dev/null
    GOOS=${os} GOARCH=${arch} go build github.com/sinbad/spriterecolour
    cp ../../README.md ./README.md
    zipfile=../SpriteRecolour-${os}-${arch}-${VERSION}.zip
    rm -f ${zipfile}
    zip ${zipfile} ./*
    popd > /dev/null
done

echo Done, see 'dist' directory for output