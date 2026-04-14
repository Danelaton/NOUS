#!/bin/bash
set -e

# NOUS Installer for macOS / Linux
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/nous-cli/nous/main/installs/install.sh | bash
#
# What this installs (all global, nothing in your projects):
#   ~/.local/bin/nous     — NOUS binary
#   ~/.nous/venv/         — Python venv with mempalace + chromadb
#   ~/.nous/hooks/        — auto-save hooks for agents
#   ~/.nous/config/       — agent configs (only for agents detected on this machine)
#
# To activate SDD in a project (run AFTER install, inside your project):
#   cd ~/my-project && nous sdd-init

GITHUB_OWNER="nous-cli"
GITHUB_REPO="nous"
BREW_TAP="nous-cli/tap"
NOUS_DIR="$HOME/.nous"
NOUS_VENV="$NOUS_DIR/venv"

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
echo -e "${C}=================================================${N}"
echo ""

# ============================================================================
# PHASE 1: Install NOUS binary
# ============================================================================
info "Phase 1/6: Installing NOUS binary..."

NOUS_INSTALLED=false

install_brew() {
    info "Installing via Homebrew..."
    brew tap "$BREW_TAP" --quiet 2>/dev/null || true
    if brew list nous &>/dev/null 2>&1; then
        brew upgrade nous --quiet 2>/dev/null || true
    else
        brew install nous --quiet
    fi
    command -v nous &>/dev/null
}

install_binary() {
    local os arch tmp
    tmp="$(mktemp -d)"
    case "$(uname -s)" in Darwin) os="darwin";; Linux) os="linux";; *) return 1;; esac
    case "$(uname -m)" in x86_64|amd64) arch="amd64";; arm64|aarch64) arch="arm64";; *) return 1;; esac

    local version
    version=$(curl -fsSL "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest" \
        2>/dev/null | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
    [ -z "$version" ] && { warn "Could not fetch latest release"; rm -rf "$tmp"; return 1; }

    local ver="${version#v}"
    local archive="nous_${ver}_${os}_${arch}.tar.gz"
    local url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${version}/${archive}"

    dim "Downloading nous ${version} (${os}/${arch})..."
    curl -fsSL "$url" -o "${tmp}/${archive}" 2>/dev/null || { warn "Download failed"; rm -rf "$tmp"; return 1; }
    tar -xzf "${tmp}/${archive}" -C "$tmp"

    local install_dir="$HOME/.local/bin"
    mkdir -p "$install_dir"
    mv "${tmp}/nous" "${install_dir}/nous" 2>/dev/null || mv "${tmp}/nous_${ver}_${os}_${arch}/nous" "${install_dir}/nous"
    chmod +x "${install_dir}/nous"
    export PATH="${install_dir}:$PATH"
    rm -rf "$tmp"
    success "nous ${version} installed to ${install_dir}"
}

install_go() {
    command -v go &>/dev/null || { err "Go not found — install from https://go.dev/dl/"; return 1; }
    GOBIN="$HOME/.local/bin" go install "github.com/${GITHUB_OWNER}/${GITHUB_REPO}/cmd/nous@latest"
    export PATH="$HOME/.local/bin:$PATH"
    command -v nous &>/dev/null
}

if command -v brew &>/dev/null; then
    install_brew && NOUS_INSTALLED=true || warn "Homebrew install failed — trying binary..."
fi
if [ "$NOUS_INSTALLED" = false ]; then
    install_binary && NOUS_INSTALLED=true || warn "Binary install failed — trying go install..."
fi
if [ "$NOUS_INSTALLED" = false ]; then
    install_go && NOUS_INSTALLED=true || { err "All install methods failed."; exit 1; }
fi

# ============================================================================
# PHASE 2: Install uv
# ============================================================================
echo ""
info "Phase 2/6: Installing uv (Python package manager)..."

UV_PATH=""
if command -v uv &>/dev/null; then
    UV_PATH="$(command -v uv)"
    dim "uv already installed: $(uv --version)"
elif [ -f "$HOME/.local/bin/uv" ]; then
    UV_PATH="$HOME/.local/bin/uv"
    dim "uv found at $UV_PATH"
else
    curl -LsSf https://astral.sh/uv/install.sh | sh
    UV_PATH="$HOME/.local/bin/uv"
fi
[ -d "$HOME/.local/bin" ] && export PATH="$HOME/.local/bin:$PATH"
"$UV_PATH" --version >/dev/null 2>&1 || { err "uv installation failed"; exit 1; }
success "uv ready: $("$UV_PATH" --version)"

# ============================================================================
# PHASE 3: Create ~/.nous/ structure
# ============================================================================
echo ""
info "Phase 3/6: Creating ~/.nous/ structure..."

mkdir -p "$NOUS_DIR/config"
mkdir -p "$NOUS_DIR/hooks"
mkdir -p "$NOUS_DIR/skills"
# Note: no mempalace/ subdir — mempalace lives inside the venv

success "~/.nous/ ready"

# ============================================================================
# PHASE 4: Create venv + install mempalace from PyPI
# ============================================================================
echo ""
info "Phase 4/6: Installing mempalace into ~/.nous/venv/ ..."

PYTHON_EXE="$NOUS_VENV/bin/python3"

# Recreate venv if python not present
if [ ! -x "$PYTHON_EXE" ]; then
    dim "Creating virtual environment..."
    "$UV_PATH" venv "$NOUS_VENV" --python python3 --quiet
fi

# Check if mempalace already installed
if "$PYTHON_EXE" -c "import mempalace" 2>/dev/null; then
    dim "mempalace already installed — skipping"
else
    dim "Running: uv pip install mempalace"
    "$UV_PATH" pip install mempalace --python "$PYTHON_EXE" --quiet
fi

# Verify
MP_VERSION=$("$PYTHON_EXE" -m mempalace --version 2>/dev/null | head -1 || echo "")
CHROMA_VERSION=$("$PYTHON_EXE" -c "import chromadb; print(chromadb.__version__)" 2>/dev/null || echo "")

if [ -n "$MP_VERSION" ]; then
    success "mempalace: $MP_VERSION"
else
    warn "mempalace CLI not responding — check manually: $PYTHON_EXE -m mempalace status"
fi
[ -n "$CHROMA_VERSION" ] && dim "chromadb: $CHROMA_VERSION"

# ============================================================================
# PHASE 5: Write auto-save hooks
# ============================================================================
echo ""
info "Phase 5/6: Writing auto-save hooks..."

cat > "$NOUS_DIR/hooks/mempal_save_hook.sh" << HOOK
#!/bin/bash
# NOUS auto-save hook — triggered by agent on Stop
"$PYTHON_EXE" -m mempalace save --checkpoint 2>/dev/null || true
HOOK

cat > "$NOUS_DIR/hooks/mempal_precompact_hook.sh" << HOOK
#!/bin/bash
# NOUS pre-compact hook — triggered by agent before context compression
"$PYTHON_EXE" -m mempalace save --emergency 2>/dev/null || true
HOOK

chmod +x "$NOUS_DIR/hooks/mempal_save_hook.sh"
chmod +x "$NOUS_DIR/hooks/mempal_precompact_hook.sh"

# PowerShell versions (for cross-platform installs or WSL)
cat > "$NOUS_DIR/hooks/mempal_save_hook.ps1" << HOOK
# NOUS auto-save hook
& "$PYTHON_EXE" -m mempalace save --checkpoint 2>\$null
HOOK
cat > "$NOUS_DIR/hooks/mempal_precompact_hook.ps1" << HOOK
# NOUS pre-compact hook
& "$PYTHON_EXE" -m mempalace save --emergency 2>\$null
HOOK

success "Hooks installed in $NOUS_DIR/hooks/"

# ============================================================================
# PHASE 6: Run nous install to inject agent configs
# ============================================================================
echo ""
info "Phase 6/6: Configuring detected agents..."

if command -v nous &>/dev/null; then
    nous install 2>/dev/null || warn "Agent config injection skipped — run 'nous sync' manually"
else
    warn "nous binary not in PATH yet — restart shell then run 'nous sync'"
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo -e "${C}=================================================${N}"
echo -e "${C}  NOUS Installation Complete${N}"
echo -e "${C}=================================================${N}"
printf "  ${G}%-20s${N} %s\n" "nous binary:" "$(command -v nous 2>/dev/null || echo 'restart shell to activate')"
printf "  ${G}%-20s${N} %s\n" "runtime dir:" "$NOUS_DIR"
printf "  ${G}%-20s${N} %s\n" "mempalace:" "${MP_VERSION:-installed}"
printf "  ${G}%-20s${N} %s\n" "chromadb:" "${CHROMA_VERSION:-installed}"
echo ""
echo -e "${C}  Next steps:${N}"
echo ""
printf "  %-25s %s\n" "cd ~/my-project" "go to any project"
printf "  %-25s %s\n" "nous sdd-init" "activate SDD workflow there (creates openspec/)"
printf "  %-25s %s\n" "nous status" "verify installation"
echo ""
echo -e "${D}  Restart your shell or run: export PATH=\"\$HOME/.local/bin:\$PATH\"${N}"
echo ""
