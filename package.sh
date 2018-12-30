version=$1
for os in linux freebsd darwin; do
    for arch in amd64; do
        make clean
        make GOOS=$os GOARCH=$arch

        pushd dist
        tar -czf ../../juun-bin-dist/juun-v$version-$os-$arch.tar.gz *
        pushd ../../juun-bin-dist/

        rm juun-latest-$os-$arch.tar.gz
        cp juun-v$version-$os-$arch.tar.gz juun-latest-$os-$arch.tar.gz
        shasum -a 256 juun-latest-$os-$arch.tar.gz
        popd
        popd
    done
done
