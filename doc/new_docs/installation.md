# Installing BetterGo from source

BetterGo is built the same way as upstream Go: a **bootstrap Go toolchain** compiles the sources in this tree, then reinstalls the result as your new `GOROOT`. The process follows [Installing Go from source](https://go.dev/doc/install/source); this document covers BetterGo-specific paths and Windows prerequisites.

## Overview

1. Install a **bootstrap** Go binary (upstream, not BetterGo) — [go.dev/dl](https://go.dev/dl/).
2. Clone or copy this repository; the `go/` directory is your future `GOROOT`.
3. Run **`make.bash`** (Linux/macOS) or **`make.bat`** (Windows) from `$GOROOT/src`.
4. Set **`GOROOT`** and **`PATH`** to use BetterGo’s `bin/go`.

| Step | Linux / macOS | Windows |
| ---- | ------------- | ------- |
| Bootstrap | Go ≥ 1.24.6 from [go.dev/dl](https://go.dev/dl/) | Same + MinGW (see below) |
| Build (toolchain only) | `./make.bash` | `make.bat` |
| Build + tests | `./all.bash` | `all.bat` |
| Run from | `$GOROOT/src` | `%GOROOT%\src` |

Do **not** use `make.bash` on Windows; use `make.bat`. See also the upstream [Windows build wiki](https://go.dev/wiki/WindowsBuild).

---

## Bootstrap toolchain

The Go compiler is written in Go. **`make.bash` / `make.bat`** need an existing `go` command to bootstrap.

- **Minimum bootstrap version:** Go **1.24.6** (see `$GOROOT/src/make.bash` / `make.bat`).
- Set **`GOROOT_BOOTSTRAP`** to the root of your upstream install if `go` on `PATH` points at BetterGo or is missing.
- **`GOROOT_BOOTSTRAP`** must **not** be the BetterGo tree you are building.

If unset, the scripts look for another `go` on `PATH`, then common locations such as `$HOME/go1.24.6` or `%USERPROFILE%\go1.24.6`.

Download bootstrap binaries: [https://go.dev/dl/](https://go.dev/dl/)  
Binary install guide: [https://go.dev/doc/install](https://go.dev/doc/install)

---

## Linux and macOS

```bash
export GOROOT_BOOTSTRAP=$(go env GOROOT)

# Optional — required for generic methods in this fork
export GOEXPERIMENT=genericmethods

cd /path/to/bettergo/go/src
./make.bash          # toolchain only
# ./all.bash         # toolchain + full test suite (long)
```

After a successful build:

```bash
export GOROOT=/path/to/bettergo/go
export PATH=$GOROOT/bin:$PATH
go version
```

Add those exports to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to keep BetterGo as the default.

### Optional: C compiler

To build with **cgo** support, install `gcc` or `clang`. To build without cgo:

```bash
CGO_ENABLED=0 ./make.bash
```

---

## Windows

Windows builds use **`make.bat`** from `%GOROOT%\src`. Upstream documents additional setup on the [Go Wiki: WindowsBuild](https://go.dev/wiki/WindowsBuild) page.

### Install MinGW (C compiler)

BetterGo’s Windows build expects a C toolchain on `PATH` (typically **`gcc.exe`**) for parts of the runtime and cgo. Install **MinGW** with the **MSYS Basic System** and ensure `gcc` is available.

**Using the MinGW installer (upstream wiki approach):**

1. Download the MinGW automated installer from [SourceForge — MinGW OldFiles](https://sourceforge.net/projects/mingw/files/OldFiles/mingw-get-inst/) (see [WindowsBuild](https://go.dev/wiki/WindowsBuild)).
2. Run the installer (`mingw-get-inst-*.exe`).
3. On **Select Components**, enable:
   - **MinGW Compiler Suite** → C Compiler
   - **MinGW Developer Toolkit**
   - **MSYS Basic System**
4. Install to a path such as `C:\MinGW`.
5. Add MinGW to **`PATH`**, for example:
   - `C:\MinGW\bin`
   - `C:\MinGW\msys\1.0\bin` (if using MSYS shell tools)

Verify:

```powershell
gcc --version
```

**64-bit (`GOARCH=amd64`):** build a working 32-bit toolchain first, then install a 64-bit MinGW-w64 toolchain and point `gcc.exe` / `ar.exe` at the 64-bit versions. Details: [WindowsBuild — 64-bit notes](https://go.dev/wiki/WindowsBuild).

**Alternative:** modern **MSYS2** / **winlibs** / **TDM-GCC** distributions also work if `gcc` is on `PATH` and matches your target `GOARCH`.

To skip cgo during the build:

```powershell
$env:CGO_ENABLED = "0"
```

### Build BetterGo

Install upstream Go from [go.dev/dl](https://go.dev/dl/) first.

```powershell
$env:GOROOT_BOOTSTRAP = (go env GOROOT)

# Optional
$env:GOEXPERIMENT = "genericmethods"

cd C:\path\to\bettergo\go\src
.\make.bat           # toolchain only
# .\all.bat          # toolchain + tests (long)
```

After a successful build:

```powershell
$env:GOROOT = "C:\path\to\bettergo\go"
$env:PATH = "$env:GOROOT\bin;$env:PATH"
go version
```

Set **`GOROOT`** and update **`PATH`** in System Environment Variables to persist across sessions.

---

## Environment variables

| Variable | Purpose |
| -------- | ------- |
| `GOROOT_BOOTSTRAP` | Upstream Go tree used to compile BetterGo (must contain `bin/go` or `bin\go.exe`) |
| `GOROOT` | Root of the **built** BetterGo tree (set after install) |
| `GOEXPERIMENT=genericmethods` | Enable generic methods (extension methods, LINQ, etc.) |
| `CGO_ENABLED=0` | Build without cgo (no C compiler required) |
| `GOOS` / `GOARCH` | Cross-compile target (defaults to host) |

---

## Verify the installation

```bash
cd $HOME
cat > hello.go <<'EOF'
package main

import "fmt"

func main() {
	fmt.Println("hello, BetterGo")
}
EOF

go run hello.go
```

---

## gopls (IDE support)

Build [gopls](doc/new_features/gopls.md) from the `go_tools` repository against BetterGo’s `GOROOT`:

```bash
export GOROOT=/path/to/bettergo/go
export GOPATH=/path/to/gopath    # must not equal GOROOT
export GOEXPERIMENT=genericmethods
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

cd /path/to/go_tools/gopls
go build -o $GOPATH/bin/gopls .
```

Configure your editor to use this `gopls` binary and the same `GOROOT`.

---

## Further reading

- [Installing Go from source (upstream)](https://go.dev/doc/install/source)
- [Go Wiki: WindowsBuild](https://go.dev/wiki/WindowsBuild)
- [Compiler testdir testing](testing.md) — running fork compiler tests after a build
