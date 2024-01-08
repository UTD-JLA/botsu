#!/usr/bin/env elvish
use flag

var extensions = [&windows=.exe &linux=]
var target-os = [windows linux]
var target-arch = [amd64]

fn build  {|pkg os arch|
    set-env GOOS $os
    set-env GOARCH $arch
    go build -o ./bin/$pkg'-'$os'-'$arch$extensions[$os] ./cmd/$pkg
}

fn build-all  {|pkg|
    for os $target-os {
        for arch $target-arch {
            build $pkg $os $arch
        }
    }
}

fn main {|&pkg=botsu| build-all $pkg}

flag:call $main~ $args