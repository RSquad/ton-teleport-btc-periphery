#!/bin/bash

echo "Setting up git hooks..."
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit

echo "Installing gofumpt code formatter..."
go install mvdan.cc/gofumpt@latest

echo "Setting up environment files..."
dirs=(./relayer ./oracle ./indexer ./metrics)

for dir in "${dirs[@]}"; do
  echo "Processing directory: $dir"
  
  if [ ! -d "$dir" ]; then
    echo "❌ Directory $dir does not exist, skipping"
    continue
  fi

  cd "$dir" || exit
  
  if [ -L ".env" ]; then
    echo "🔄 Found existing symlink in $dir, removing it"
    rm .env
  elif [ -e ".env" ]; then
    echo "⚠️ Regular .env file exists in $dir, skipping"
    cd - || exit
    continue
  fi
  
  echo "✅ Creating symbolic link to .env in $dir"
  ln -s ../.env .env
  cd - || exit
done

echo "Setup complete! 🎉"
