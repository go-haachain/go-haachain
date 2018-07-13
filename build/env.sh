#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
haadir="$workspace/src/github.com/haaereum"
if [ ! -L "$haadir/go-haaereum" ]; then
    mkdir -p "$haadir"
    cd "$haadir"
    ln -s ../../../../../. go-haaereum
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$haadir/go-haaereum"
PWD="$haadir/go-haaereum"

# Launch the arguments with the configured environment.
exec "$@"
