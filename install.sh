#!/usr/bin/env bash
set -e

# FilmTag Installer Script

# Colors for output
green="\033[0;32m"
yellow="\033[1;33m"
red="\033[0;31m"
reset="\033[0m"

# Check for Go
if ! command -v go >/dev/null 2>&1; then
  echo -e "${red}Error: Go is not installed. Please install Go (https://golang.org/dl/) and try again.${reset}"
  exit 1
fi

# Check for exiftool
if ! command -v exiftool >/dev/null 2>&1; then
  echo -e "${yellow}ExifTool not found. Attempting to install...${reset}"
  OS_NAME=$(uname)
  if [ "$OS_NAME" = "Darwin" ]; then
    if command -v brew >/dev/null 2>&1; then
      echo -e "${green}Installing exiftool via Homebrew...${reset}"
      brew install exiftool
    else
      echo -e "${red}Homebrew is not installed. Please install Homebrew (https://brew.sh/) and then run this script again, or install exiftool manually.${reset}"
      exit 1
    fi
  else
    echo -e "${red}ExifTool is required but was not found. Please install exiftool using your system's package manager and try again.${reset}"
    exit 1
  fi
fi

# Install filmtag
echo -e "${green}Installing filmtag...${reset}"
go install github.com/rohanpandula/filmtag@latest

# Get Go bin path
GOPATH=$(go env GOPATH)
GOBIN="$GOPATH/bin"

# Detect shell and config file
SHELL_NAME=$(basename "$SHELL")
CONFIG_FILE=""
case "$SHELL_NAME" in
  zsh)
    CONFIG_FILE="$HOME/.zshrc"
    ;;
  bash)
    if [ -f "$HOME/.bash_profile" ]; then
      CONFIG_FILE="$HOME/.bash_profile"
    else
      CONFIG_FILE="$HOME/.bashrc"
    fi
    ;;
  fish)
    CONFIG_FILE="$HOME/.config/fish/config.fish"
    ;;
  *)
    CONFIG_FILE="$HOME/.profile"
    ;;
esac

# Check if GOBIN is already in PATH in config file
if grep -qs "export PATH=\"\$PATH:$GOBIN\"" "$CONFIG_FILE"; then
  echo -e "${yellow}Go bin path already present in $CONFIG_FILE${reset}"
else
  echo -e "${green}Adding Go bin path to $CONFIG_FILE...${reset}"
  echo "export PATH=\"\$PATH:$GOBIN\"" >> "$CONFIG_FILE"
fi

# Final message
echo -e "${green}Installation complete!${reset}"
echo -e "\nTo use filmtag, restart your terminal or run:"
echo -e "  source $CONFIG_FILE"
echo -e "\nYou can now run: filmtag --help" 