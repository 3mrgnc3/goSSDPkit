#!/bin/bash

# goSSDPkit Cross-Platform Release Builder
# Clean, colorful cross-compilation script

set -e

# Beautiful terminal colors
readonly GREEN='\033[38;5;46m'
readonly BLUE='\033[38;5;51m' 
readonly PURPLE='\033[38;5;129m'
readonly ORANGE='\033[38;5;208m'
readonly CYAN='\033[38;5;87m'
readonly YELLOW='\033[38;5;226m'
readonly RED='\033[38;5;196m'
readonly GRAY='\033[38;5;244m'
readonly BOLD='\033[1m'
readonly NC='\033[0m'

# Configuration
VERSION=${1:-"dev"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Help option
if [[ $1 == "-h" || $1 == "--help" ]]; then
    echo -e "${BLUE}${BOLD}goSSDPkit Cross-Platform Release Builder${NC}"
    echo
    echo "Usage: $0 [VERSION]"
    echo
    echo "Arguments:"
    echo "  VERSION    Version string (default: 'dev')"
    echo
    echo "Examples:"
    echo "  $0              # Build with version 'dev'"
    echo "  $0 v1.0.0       # Build with version 'v1.0.0'"
    echo "  $0 \$(git describe --tags)  # Use current git tag"
    echo
    echo "Builds 17 platform binaries with embedded version info:"
    echo "  • Desktop: Linux, Windows, macOS (x64 + ARM)"
    echo "  • Embedded: ARM64, ARMv7, MIPS, MIPSLE" 
    echo "  • Server: PowerPC, S390X, RISC-V, BSD variants"
    exit 0
fi

# Print functions
print_header() { echo -e "${BLUE}${BOLD}$1${NC}"; }
print_info() { echo -e "${CYAN}→ $1${NC}"; }
print_success() { echo -e "${GREEN}✓ $1 ${GRAY}($(du -h $2 | cut -f1))${NC}"; }

# Header
print_header "🚀 goSSDPkit Cross-Platform Release Builder"
echo
echo -e "${PURPLE}Version:${NC} $VERSION"
echo -e "${PURPLE}Build Time:${NC} $BUILD_TIME" 
echo -e "${PURPLE}Git Commit:${NC} $GIT_COMMIT"
echo

# Prepare build directory
rm -rf releases
mkdir -p releases

# Build function
build() {
    local name="$1" goos="$2" goarch="$3" goarm="$4" output="$5"
    
    print_info "Building $name"
    
    # Set environment
    export GOOS="$goos" GOARCH="$goarch"
    [[ -n "$goarm" ]] && export GOARM="$goarm" || unset GOARM
    
    # Build with ldflags
    local ldflags="-w -s -X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME' -X 'main.GitCommit=$GIT_COMMIT'"
    
    if go build -ldflags="$ldflags" -o "releases/$output" ./cmd/goSSDPkit; then
        print_success "$name" "releases/$output"
    else
        echo -e "${RED}✗ Failed: $name${NC}"
    fi
    
    unset GOOS GOARCH GOARM
}

# Core Desktop Platforms
print_header "🖥️  Desktop Platforms"
build "Linux AMD64" "linux" "amd64" "" "goSSDPkit_linux_amd64"
build "Windows AMD64" "windows" "amd64" "" "goSSDPkit_windows_amd64.exe" 
build "macOS AMD64" "darwin" "amd64" "" "goSSDPkit_darwin_amd64"
build "macOS ARM64 (Apple Silicon)" "darwin" "arm64" "" "goSSDPkit_darwin_arm64"
echo

# ARM & Embedded Platforms
print_header "🔧 ARM & Embedded Platforms"
build "Linux ARM64" "linux" "arm64" "" "goSSDPkit_linux_arm64"
build "Linux ARMv7 (Raspberry Pi)" "linux" "arm" "7" "goSSDPkit_linux_armv7"
build "Linux MIPSLE (Routers)" "linux" "mipsle" "" "goSSDPkit_linux_mipsle"
build "Linux MIPS" "linux" "mips" "" "goSSDPkit_linux_mips"
echo

# Extended Platforms
print_header "🌐 Extended Platforms"
build "Linux 386" "linux" "386" "" "goSSDPkit_linux_386"
build "Windows 386" "windows" "386" "" "goSSDPkit_windows_386.exe"
build "Windows ARM64" "windows" "arm64" "" "goSSDPkit_windows_arm64.exe"
build "FreeBSD AMD64" "freebsd" "amd64" "" "goSSDPkit_freebsd_amd64"
build "OpenBSD AMD64" "openbsd" "amd64" "" "goSSDPkit_openbsd_amd64"
echo

# Server Platforms  
print_header "🏢 Server & Enterprise Platforms"
build "Linux MIPS64LE" "linux" "mips64le" "" "goSSDPkit_linux_mips64le"
build "Linux PPC64LE (IBM Power)" "linux" "ppc64le" "" "goSSDPkit_linux_ppc64le"
build "Linux S390X (IBM Z)" "linux" "s390x" "" "goSSDPkit_linux_s390x"
build "Linux RISC-V 64" "linux" "riscv64" "" "goSSDPkit_linux_riscv64"
echo

# Summary
print_header "📊 Build Summary"
total_size=$(du -sh releases | cut -f1)
file_count=$(ls releases/ | wc -l)

echo -e "${GREEN}✅ Successfully built $file_count platform binaries${NC}"
echo -e "${YELLOW}📁 Total size: $total_size${NC}"
echo -e "${CYAN}📦 Files created in: releases/${NC}"
echo

# List files with colors
echo -e "${GRAY}Release files:${NC}"
ls -lah releases/ | tail -n +2 | while read -r line; do
    echo -e "${GRAY}  $line${NC}"
done

echo
print_header "🎉 Cross-compilation complete!"