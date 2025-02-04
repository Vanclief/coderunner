#!/usr/bin/env bash
set -e

echo "Installing coderunner..."

TARGET_DIR="$HOME/.coderunner/bin"
TARGET_DIR_BIN="$TARGET_DIR/coderunner"

SERVER="https://raw.githubusercontent.com/vanclief/coderunner/master/bin"

# Detect the platform (similar to $OSTYPE)
OS="$(uname)"
if [[ "$OS" == "Linux" ]]; then
    # Linux
    FILENAME="coderunner-linux"
elif [[ "$OS" == "Darwin" ]]; then
    # MacOS, should validate if Intel or ARM
    UNAMEM="$(uname -m)"
    if [[ "$UNAMEM" == "x86_64" ]]; then
        FILENAME="coderunner-mac-amd64"
    else
        FILENAME="coderunner-mac-arm64"
    fi
else
    echo "unrecognized OS: $OS"
    echo "Exiting..."
    exit 1
fi

# Check if ~/.coderunner/bin exists, if not create it
if [[ ! -e "${TARGET_DIR}" ]]; then
    mkdir -p "${TARGET_DIR}"
fi

# Download the appropriate binary
echo "Downloading $SERVER/$FILENAME..."
curl -# -L "${SERVER}/${FILENAME}" -o "${TARGET_DIR_BIN}"
chmod +x "${TARGET_DIR_BIN}"
echo "Installed under ${TARGET_DIR_BIN}"

# Store the correct profile file (i.e. .profile for bash or .zshenv for ZSH).
case $SHELL in
*/zsh)
    PROFILE=${ZDOTDIR-"$HOME"}/.zshenv
    PREF_SHELL=zsh
    APPEND_COMMAND="export PATH=\"\$PATH:$TARGET_DIR\""
    ;;
*/bash)
    PROFILE=$HOME/.bashrc
    PREF_SHELL=bash
    APPEND_COMMAND="export PATH=\"\$PATH:$TARGET_DIR\""
    ;;
*/fish)
    PROFILE=$HOME/.config/fish/config.fish
    PREF_SHELL=fish
    APPEND_COMMAND="set -U fish_user_paths \$fish_user_paths $TARGET_DIR"
    ;;
*/ash)
    PROFILE=$HOME/.profile
    PREF_SHELL=ash
    APPEND_COMMAND="export PATH=\"\$PATH:$TARGET_DIR\""
    ;;
*)
    echo "could not detect shell, manually add ${TARGET_DIR} to your PATH."
    exit 1
    ;;
esac

# Only add if it isn't already in PATH.
if [[ ":$PATH:" != *":${TARGET_DIR}:"* ]]; then
    echo >>"$PROFILE"
    echo "$APPEND_COMMAND" >>"$PROFILE"
fi

# Reload the profile
if [[ "$PREF_SHELL" == "fish" ]]; then
    # For fish, we need to reload the fish_user_paths
    fish -c "source $PROFILE"
else
    # For other shells, source the profile
    source "$PROFILE"
fi

echo
echo "Detected your preferred shell is ${PREF_SHELL} and added coderunner to PATH. You can now run the command coderunner."
echo "coderunner successfully installed!"
