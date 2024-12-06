#!/bin/bash

cargo build
cbindgen --config cbindgen.toml --crate frost --output ./frost.h