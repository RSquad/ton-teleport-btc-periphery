#!/bin/bash

cargo build --release
cbindgen --config cbindgen.toml --crate frost --output ./frost.h