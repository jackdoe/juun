for os in linux freebsd darwin; do
    for arch in amd64; do
        make clean
        make GOOS=$os GOARCH=$arch

        pushd dist
        tar -czf ../../juun-bin-dist/juun-latest-$os-$arch.tar.gz *
        popd
    done
done
                
                
