#!/bin/bash
#
# This script sets up a Go work area for the `goverdrive` project,
# starting from an empty base project directory. For example, if
# you put each of your go projects under ~/proj/go/, this script
# could be run frome th sub-directory that houses the project:
#    $ cd ~/proj/go
#    $ mkdir goverdrive
#    $ cp <this_file> goverdrive
#    $ chmod +x goverdrive/proj_setup.sh
#    $ cd goverdrive
#    $ ./proj_setup.sh


######################################################################
# Setup the Go project area
######################################################################
mkdir bin pkg src

echo '
export GOENV=verdrive
export GOPATH=~/proj/go/verdrive
export PATH=~/proj/go/verdrive/bin:$PATH
' >> bin/activate

chmod +x bin/activate
. bin/activate


######################################################################
# Get latest source code for goverdrive
######################################################################
go get github.com/anki/goverdrive


######################################################################
# Get source code for supporting code
# TODO: Use a loop, for God's sake!!
######################################################################
go get github.com/faiface/glhf
pushd src/github.com/faiface/glhf
git checkout --quiet 98c0391c0fd3f0b365cfe5d467ac162b79dfb002
popd

go get github.com/faiface/mainthread
pushd src/github.com/faiface/mainthread
git checkout --quiet 7127dc8993886d1ce799086ff298ff99bf258fbf
popd

go get github.com/faiface/pixel
pushd src/github.com/faiface/pixel
git checkout --quiet 4792a9ebd80d8ed576dff53e6d9d8a761826d82e
popd

go get github.com/go-gl/gl/v3.3-core/gl
pushd src/github.com/go-gl/gl/v3.3-core/gl
git checkout --quiet ac0d3d2af0fe995e2bb09fa6619b716b0c0fc0fa
popd

go get github.com/go-gl/glfw/v3.2/glfw
pushd src/github.com/go-gl/glfw/v3.2/glfw
git checkout --quiet 513e4f2bf85c31fba0fc4907abd7895242ccbe50
popd

go get github.com/go-gl/mathgl/mgl32
pushd src/github.com/go-gl/mathgl/mgl32
git checkout --quiet 4c3fc6b4bf30179013667e5d0b926e024d9a13c1
popd

go get github.com/pkg/errors
pushd src/github.com/pkg/errors
git checkout --quiet 2b3a18b5f0fb6b4f9190549597d3f962c02bc5eb
popd

go get golang.org/x/image/colornames
pushd src/golang.org/x/image/colornames
git checkout --quiet e20db36d77bd0cb36cea8fe49d5c37d82d21591f
popd

go get golang.org/x/image/font
pushd src/golang.org/x/image/font
git checkout --quiet e20db36d77bd0cb36cea8fe49d5c37d82d21591f
popd

go get golang.org/x/image/math/f32
pushd src/golang.org/x/image/math/f32
git checkout --quiet e20db36d77bd0cb36cea8fe49d5c37d82d21591f
popd

go get golang.org/x/image/math/fixed
pushd src/golang.org/x/image/math/fixed
git checkout --quiet e20db36d77bd0cb36cea8fe49d5c37d82d21591f
popd
