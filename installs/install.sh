#!/bin/bash
set -e

# NOUS Installer for macOS / Linux
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.sh | bash
#
# What this installs:
#   ~/.local/bin/nous              — NOUS binary
#   ~/.nous/skills/                — skills (AGENTS.md)
#   ~/.nous/config/                — agent configs
#
# To activate a project:
#   cd ~/my-project && nous sdd-init
#   cd ~/my-project && nous sync

GITHUB_OWNER="Danelaton"
GITHUB_REPO="NOUS"

# Auto-detect latest tag from GitHub API
VERSION=$(curl -fsSL "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest" 2>/dev/null \
    | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$VERSION" ]; then
    printf "[NOUS] Could not fetch latest release — using default version\n"
    VERSION="v2026.4.14-1"
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
printf "${C}  NOUS — AI Ecosystem Configurator${N}\n"
printf "${C}  Version: ${VERSION}${N}\n"
printf "${C}=================================================${N}\n"
printf "\n"

# ============================================================================
# PHASE 1: Install NOUS binary
# ============================================================================
info "Phase 1/5: Installing NOUS binary..."

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
# PHASE 2: Install skills from GitHub
# ============================================================================
printf "\n"
info "Phase 2/5: Installing skills..."

mkdir -p "$SKILLS_DIR"
AGENTS_URL="https://raw.githubusercontent.com/${GITHUB_OWNER}/${GITHUB_REPO}/${VERSION}/installs/skeleton/AGENTS.md"
if curl -fsSL "$AGENTS_URL" -o "$SKILLS_DIR/AGENTS.md" 2>/dev/null; then
    success "AGENTS.md installed"
else
    warn "Could not download AGENTS.md — skipping skills"
fi

# Download skills folders from installs/skills/
install_skills_folder() {
    local repo_path="$1"
    local dest_dir="$2"
    local api_url="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/contents/${repo_path}"
    local items
    items=$(curl -fsSL "$api_url" 2>/dev/null) || return 0
    echo "$items" | while IFS= read -r item; do
        # Parse name and type from JSON — using grep/sed for portability
        name=$(echo "$item" | grep '"name"' | sed 's/.*"name": *"\([^"]*\)".*/\1/' | head -1)
        type=$(echo "$item" | grep '"type"' | sed 's/.*"type": *"\([^"]*\)".*/\1/' | head -1)
        download_url=$(echo "$item" | grep '"download_url"' | sed 's/.*"download_url": *"\([^"]*\)".*/\1/' | head -1)
        if [ -z "$name" ] || [ -z "$type" ]; then continue; fi
        if [ "$type" = "dir" ]; then
            mkdir -p "${dest_dir}/${name}"
            install_skills_folder "${repo_path}/${name}" "${dest_dir}/${name}"
        elif [ "$type" = "file" ] && [ -n "$download_url" ]; then
            curl -fsSL "$download_url" -o "${dest_dir}/${name}" 2>/dev/null
        fi
    done
}

install_skills_folder "installs/skills" "$SKILLS_DIR"
success "Skills folder installed"

# ============================================================================
# PHASE 3: Create ~/.nous/ structure
# ============================================================================
printf "\n"
info "Phase 3/5: Creating ~/.nous/ structure..."

mkdir -p "$NOUS_DIR/config"
mkdir -p "$NOUS_DIR/skills"

if [ -f "$SKILLS_DIR/AGENTS.md" ]; then
    cp "$SKILLS_DIR/AGENTS.md" "$NOUS_DIR/skills/AGENTS.md"
fi

# Copy skill folders (e.g. skill-creator/) to ~/.nous/skills/
for dir in "$SKILLS_DIR"/*/; do
    if [ -d "$dir" ]; then
        skill_name=$(basename "$dir")
        dest="$NOUS_DIR/skills/$skill_name"
        rm -rf "$dest"
        cp -r "$dir" "$dest"
    fi
done

success "~/.nous/ ready"

# ============================================================================
# PHASE 4: Run nous install (detect + inject agent configs)
# ============================================================================
printf "\n"
info "Phase 4/5: Detecting agents and configuring..."

if command -v nous &>/dev/null; then
    nous install 2>/dev/null || warn "Agent configuration skipped — run 'nous sync' manually"
else
    warn "nous binary not in PATH — restart shell then run 'nous sync'"
fi

# ============================================================================
# PHASE 5: Summary
# ============================================================================
printf "\n"
printf "${C}=================================================${N}\n"
printf "${C}  NOUS Installation Complete${N}\n"
printf "${C}  Version: ${VERSION}${N}\n"
printf "${C}=================================================${N}\n"
printf "  ${G}%-20s${N} %s\n" "nous binary:" "$(command -v nous 2>/dev/null || echo 'restart shell to activate')"
printf "  ${G}%-20s${N} %s\n" "skills:" "$SKILLS_DIR"
printf "  ${G}%-20s${N} %s\n" "config dir:" "$NOUS_DIR/config"
printf "\n"
printf "${C}  Next steps:${N}\n"
printf "\n"
printf "  %-25s %s\n" "cd ~/my-project" "go to any project"
printf "  %-25s %s\n" "nous sync" "setup dev/ + skills + AGENTS.md in project"
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
