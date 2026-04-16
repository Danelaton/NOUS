#!/bin/bash
set -e

# NOUS Installer for macOS / Linux
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Danelaton/NOUS/main/installs/install.sh | bash
#
# What this installs:
#   ~/.local/bin/nous              — NOUS binary
#   ~/.local/share/nous/skills/     — skills (AGENTS.md)
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
    echo "[NOUS] Could not fetch latest release — using default version"
    VERSION="v2026.4.14"
fi

SKILLS_DIR="$HOME/.local/share/nous/skills"
NOUS_DIR="$HOME/.nous"

# ── Colors ────────────────────────────────────────────────────────────────────
R='\033[0;31m' Y='\033[1;33m' G='\033[0;32m' C='\033[0;36m' D='\033[0;90m' N='\033[0m'
info()    { echo -e "${C}[NOUS]${N} $*"; }
success() { echo -e "${G}[NOUS]${N} $*"; }
warn()    { echo -e "${Y}[NOUS]${N} $*"; }
err()     { echo -e "${R}[NOUS]${N} $*" >&2; }
dim()     { echo -e "${D}[NOUS]${N} $*"; }

echo ""
echo -e "${C}=================================================${N}"
echo -e "${C}  NOUS — AI Ecosystem Configurator${N}"
echo -e "${C}  Version: ${VERSION}${N}"
echo -e "${C}=================================================${N}"
echo ""

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
    curl -fsSL "$url" -o "${tmp}/${archive}" || { warn "Download failed"; rm -rf "$tmp"; return 1; }
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
    success "nous ${VERSION} installed to ${install_dir}"
}

install_go() {
    command -v go &>/dev/null || { err "Go not found — install from https://go.dev/dl/"; return 1; }
    GOBIN="$HOME/.local/bin" go install "github.com/${GITHUB_OWNER}/${GITHUB_REPO}/cmd/nous@${VERSION}"
    export PATH="$HOME/.local/bin:$PATH"
    command -v nous &>/dev/null
}

install_binary && success "Binary installed" || install_go && success "Installed via go install" || { err "All install methods failed."; exit 1; }

# ============================================================================
# PHASE 2: Install skills from GitHub
# ============================================================================
echo ""
info "Phase 2/5: Installing skills..."

mkdir -p "$SKILLS_DIR"
AGENTS_URL="https://raw.githubusercontent.com/${GITHUB_OWNER}/${GITHUB_REPO}/${VERSION}/installs/skeleton/AGENTS.md"
if curl -fsSL "$AGENTS_URL" -o "$SKILLS_DIR/AGENTS.md" 2>/dev/null; then
    success "AGENTS.md installed"
else
    warn "Could not download AGENTS.md — skipping skills"
fi

# ============================================================================
# PHASE 3: Create ~/.nous/ structure
# ============================================================================
echo ""
info "Phase 3/5: Creating ~/.nous/ structure..."

mkdir -p "$NOUS_DIR/config"
mkdir -p "$NOUS_DIR/skills"

if [ -f "$SKILLS_DIR/AGENTS.md" ]; then
    cp "$SKILLS_DIR/AGENTS.md" "$NOUS_DIR/skills/AGENTS.md"
fi

success "~/.nous/ ready"

# ============================================================================
# PHASE 4: Run nous install (detect + inject agent configs)
# ============================================================================
echo ""
info "Phase 4/5: Detecting agents and configuring..."

if command -v nous &>/dev/null; then
    nous install 2>/dev/null || warn "Agent configuration skipped — run 'nous sync' manually"
else
    warn "nous binary not in PATH — restart shell then run 'nous sync'"
fi

# ============================================================================
# PHASE 5: Summary
# ============================================================================
echo ""
echo -e "${C}=================================================${N}"
echo -e "${C}  NOUS Installation Complete${N}"
echo -e "${C}  Version: ${VERSION}${N}"
echo -e "${C}=================================================${N}"
printf "  ${G}%-20s${N} %s\n" "nous binary:" "$(command -v nous 2>/dev/null || echo 'restart shell to activate')"
printf "  ${G}%-20s${N} %s\n" "skills:" "$SKILLS_DIR"
printf "  ${G}%-20s${N} %s\n" "config dir:" "$NOUS_DIR/config"
echo ""
echo -e "${C}  Next steps:${N}"
echo ""
printf "  %-25s %s\n" "cd ~/my-project" "go to any project"
printf "  %-25s %s\n" "nous sdd-init" "create openspec/ (SDD workflow)"
printf "  %-25s %s\n" "nous sync" "setup dev/ + install AGENTS.md in project"
echo ""
echo -e "${D}  Restart your shell or run: export PATH=\"\$HOME/.local/bin:\$PATH\"${N}"
echo ""
