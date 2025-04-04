#!/bin/bash

#cargo build --release
cargo build
cbindgen --config cbindgen.toml --crate frost --output ./frost.h
