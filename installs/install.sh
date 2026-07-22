#!/bin/bash
set -e

# NOUS Installer for macOS / Linux
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.sh | bash
#
# What this installs:
#   ~/.local/bin/nous              — NOUS binary
#   ~/.nous/skills/                — predefined skills
#
# To use:
#   nous sync        # setup project (dev/ + .agents/OKF/ + AGENTS.md + skills)
#   nous skills      # install skills into current project

GITHUB_OWNER="Danelaton"
GITHUB_REPO="NOUS"

# Auto-detect latest tag from GitHub API
VERSION=$(curl -fsSL "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest" 2>/dev/null \
    | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$VERSION" ]; then
    printf "[NOUS] Could not fetch latest release — using default version\n"
    VERSION="v2026.5.19"
fi

SKILLS_DIR="$HOME/.nous/skills"
NOUS_DIR="$HOME/.nous"

# ── Colors (using printf for macOS compatibility) ────────────────────────────
R='\033[0;31m' Y='\033[1;33m' G='\033[0;32m' C='\033[0;36m' D='\033[0;90m' N='\033[0m'
info()    { printf "${C}[NOUS]${N} %s\n" "$*"; }
success() { printf "${G}[NOUS]${N} %s\n" "$*"; }
warn()    { printf "${Y}[NOUS]${N} %s\n" "$*"; }
err()     { printf "${R}[NOUS]${N} %s\n" "$*" >&2; }
dim()     { printf "${D}[NOUS]${N} %s\n" "$*"; }

printf "\n"
printf "${C}=================================================${N}\n"
printf "${C}  NOUS — AI Skills Installer${N}\n"
printf "${C}  Version: ${VERSION}${N}\n"
printf "${C}=================================================${N}\n"
printf "\n"

# ============================================================================
# PHASE 1: Remove previous installation (clean upgrade)
# ============================================================================
info "Phase 1/4: Removing previous installation..."

for prev_path in "/usr/local/bin/nous" "$HOME/.local/bin/nous"; do
    if [ -f "$prev_path" ]; then
        rm -f "$prev_path"
        dim "Previous binary removed: $prev_path"
    fi
done

# ============================================================================
# PHASE 2: Install NOUS binary
# ============================================================================
printf "\n"
info "Phase 2/4: Installing NOUS binary..."

install_binary() {
    local os arch tmp
    tmp="$(mktemp -d)"
    case "$(uname -s)" in Darwin) os="darwin";; Linux) os="linux";; *) return 1;; esac
    case "$(uname -m)" in x86_64|amd64) arch="amd64";; arm64|aarch64) arch="arm64";; *) return 1;; esac

    local archive="nous_${VERSION#v}_${os}_${arch}.tar.gz"
    local url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${VERSION}/${archive}"

    dim "Downloading nous ${VERSION} (${os}/${arch})..."
    if ! curl -fsSL "$url" -o "${tmp}/${archive}"; then
        warn "Download failed"
        rm -rf "$tmp"
        return 1
    fi
    tar -xzf "${tmp}/${archive}" -C "$tmp"

    local install_dir="/usr/local/bin"
    if [ ! -w "$install_dir" ]; then
        install_dir="$HOME/.local/bin"
    fi
    mkdir -p "$install_dir"
    if [ -x "${tmp}/nous" ]; then
        mv "${tmp}/nous" "${install_dir}/nous"
    else
        find "${tmp}" -name nous -type f -exec mv {} "${install_dir}/nous" \;
    fi
    chmod +x "${install_dir}/nous"
    export PATH="${install_dir}:$PATH"
    rm -rf "$tmp"
    return 0
}

install_go() {
    if ! command -v go &>/dev/null; then
        err "Go not found — install from https://go.dev/dl/"
        return 1
    fi
    GOBIN="$HOME/.local/bin" go install "github.com/Danelaton/NOUS/cmd/nous@${VERSION}"
    export PATH="$HOME/.local/bin:$PATH"
    command -v nous &>/dev/null
}

NOUS_INSTALLED=false
if install_binary; then
    success "Binary installed"
    NOUS_INSTALLED=true
elif install_go; then
    success "Installed via go install"
    NOUS_INSTALLED=true
fi

if [ "$NOUS_INSTALLED" = false ]; then
    err "All install methods failed."
    err "Download manually from: https://github.com/Danelaton/NOUS/releases"
    exit 1
fi

# ============================================================================
# PHASE 3: Download skills to ~/.nous/skills/
# ============================================================================
printf "\n"
info "Phase 3/4: Downloading skills to ~/.nous/skills/..."

mkdir -p "$SKILLS_DIR"

# Download AGENTS.md
AGENTS_URL="https://raw.githubusercontent.com/${GITHUB_OWNER}/${GITHUB_REPO}/main/installs/skeleton/AGENTS.md"
if curl -fsSL "$AGENTS_URL" -o "$SKILLS_DIR/AGENTS.md" 2>/dev/null; then
    success "AGENTS.md downloaded"
else
    warn "Could not download AGENTS.md — skipping"
fi

# Download skills folders from installs/skills/
# Uses jq if available, falls back to python3, then python — all handle multi-line JSON
parse_github_contents() {
    local json="$1"
    if command -v jq &>/dev/null; then
        echo "$json" | jq -r '.[] | "\(.type)\t\(.name)\t\(.download_url // "")"'
    elif command -v python3 &>/dev/null; then
        echo "$json" | python3 -c "
import sys, json
items = json.load(sys.stdin)
for i in items:
    print(i['type'] + '\t' + i['name'] + '\t' + (i.get('download_url') or ''))
"
    elif command -v python &>/dev/null; then
        echo "$json" | python -c "
import sys, json
items = json.load(sys.stdin)
for i in items:
    print(i['type'] + '\t' + i['name'] + '\t' + (i.get('download_url') or ''))
"
    else
        warn "jq and python not found — skills folder download skipped"
        return 1
    fi
}

install_skills_folder() {
    local repo_path="$1"
    local dest_dir="$2"
    local api_url="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/contents/${repo_path}"
    local items
    items=$(curl -fsSL "$api_url" 2>/dev/null) || return 0
    local parsed
    parsed=$(parse_github_contents "$items") || return 0
    while IFS=$'\t' read -r type name download_url; do
        [ -z "$name" ] && continue
        if [ "$type" = "dir" ]; then
            mkdir -p "${dest_dir}/${name}"
            install_skills_folder "${repo_path}/${name}" "${dest_dir}/${name}"
        elif [ "$type" = "file" ] && [ -n "$download_url" ]; then
            curl -fsSL "$download_url" -o "${dest_dir}/${name}" 2>/dev/null
        fi
    done <<EOF
$parsed
EOF
}

install_skills_folder "installs/skills" "$SKILLS_DIR"
success "Skills downloaded to $SKILLS_DIR"

# ============================================================================
# PHASE 4: Summary
# ============================================================================
printf "\n"
printf "${C}=================================================${N}\n"
printf "${C}  Installation Complete${N}\n"
printf "${C}  Version: ${VERSION}${N}\n"
printf "${C}=================================================${N}\n"
printf "  ${G}%-20s${N} %s\n" "nous binary:" "$(command -v nous 2>/dev/null || echo 'restart shell to activate')"
printf "  ${G}%-20s${N} %s\n" "skills:" "$SKILLS_DIR"
printf "\n"
printf "${C}  Usage:${N}\n"
printf "\n"
printf "  %-30s %s\n" "cd ~/my-project" ""
printf "  %-30s %s\n" "nous sync" "setup project: dev/ + .agents/OKF/ + AGENTS.md + skills"
printf "  %-30s %s\n" "nous skills" "install/update skills in current project"
printf "\n"

# ── PATH hint: detect shell and show correct rc file ──────────────────────
if [ -n "$ZSH_VERSION" ]; then
    RC="~/.zshrc"
elif [ -n "$BASH_VERSION" ]; then
    RC="~/.bash_profile"
else
    RC="~/.profile"
fi
printf "  ${D}Add to %s: export PATH=\"\$HOME/.local/bin:\$PATH\"${N}\n" "$RC"
printf "\n"
