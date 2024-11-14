#!/bin/bash

dirs=(./relayer ./oracle)

for dir in "${dirs[@]}"; do
  if [ -d "$dir" ]; then
    cd "$dir" || exit
    if [ -L ".env" ]; then
      echo "Removing existing symbolic link in $dir"
      rm .env
    elif [ -e ".env" ]; then
      echo "A regular file named .env exists in $dir. Skipping."
      cd - || exit
      continue
    fi
    echo "Creating symbolic link in $dir"
    ln -s ../.env .env
    cd - || exit
  else
    echo "Directory $dir does not exist."
  fi
done
