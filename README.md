# 🖼️ Sharpy (Go)

Go port of the [sharpy-cli-tool](https://github.com/codegeekery/sharpy-cli-tool) CLI (originally in Node.js + [Sharp](https://sharp.pixelplumbing.com/)).
Bulk convert images between modern formats with the exact same command-line interface, options, and behavior as the original version.

## Why `govips`?

Sharp (the Node.js library) is essentially a wrapper around **libvips**. To maintain the exact same quality and set of formats (including AVIF),
this port uses [`govips`](https://github.com/davidbyttow/govips), which provides Go bindings for libvips—the exact same underlying engine. The result is functionally equivalent to the original.

## Requirements

Before anything else, you need **Go** installed on your system (Go 1.21 or newer is recommended).
You can check with:

```bash
go version
```

If Go is not installed, download it from [go.dev/dl](https://go.dev/dl/) and follow the
instructions for your OS.

In addition, since `govips` uses cgo on top of libvips, you also need libvips installed on your
system (with AVIF/HEIF support via libheif):

```bash
# Debian/Ubuntu
sudo apt-get install -y libvips-dev libheif-plugin-aomenc pkg-config gcc

# macOS
brew install vips

# Windows (via MSYS2)
pacman -S mingw-w64-x86_64-toolchain mingw-w64-x86_64-pkg-config mingw-w64-x86_64-libvips
```

`libheif-plugin-aomenc` is specifically required to **encode** AVIF (without it, converting to AVIF
fails with "Unsupported compression"; reading/decoding AVIF works fine without it).



## Installation

### Option 1: Download the binary (recommended)

Grab the binary for your OS from the [Releases](../../releases) page, then make it executable (Linux/macOS):

```bash
chmod +x sharpy-linux-amd64
```

Remember: you still need **libvips** installed on your system for the binary to run (see [Requirements](#requirements) above).

### Option 2: Build from source

If you'd rather build it yourself, clone the repository and build the binary:

```bash
git clone <repo-url>
cd sharpy-go
go mod tidy
go build -o sharpy .
```

> **Windows:** build from within an MSYS2 MINGW64 shell (see [Requirements](#requirements)) so the
> C compiler and libvips are correctly linked.

Optional: install it into `$GOPATH/bin` so you can run `sharpy` from anywhere:

```bash
go install .
```


## Usage

Same as the original:

```bash
sharpy <format> [options]
```

Formats: `jpeg` · `jpg` · `png` · `webp` · `avif` · `tiff`

### Options

| Option                | Alias  | Description                                                |
| --------------------- | ------ | ------------------------------------------------------------ |
| `--dir <path>`         |        | Folder to process (default: current folder)                |
| `--recursive`          | `-r`   | Process images in subfolders                                |
| `--quality <n>`        | `-q`   | Quality (0-100) for formats with compression                 |
| `--force`              | `-f`   | Overwrite existing destination files                         |
| `--remove-original`    | `--rm` | Delete the original file after a successful conversion       |
| `--rename <strategy>`  |        | Naming strategy: `original` (default), `snowflake`, `uuid`  |
| `--help`               | `-h`   | Show help                                                    |

### Examples

```bash
sharpy webp
sharpy avif -q 70 -r --rm
sharpy jpeg --dir ./photos -f -q 85
sharpy png -r --dir ./images
sharpy webp --rename snowflake
sharpy avif --rename uuid -r
```

## Differences from the original

- **Concurrency**: same as the original (4 concurrent workers), implemented with goroutines +
  `sync/atomic` instead of Node.js promises.
- **Snowflake/UUID**: same algorithm (timestamp << 22 | workerId | 12-bit sequence) and UUID v4
  generated with `crypto/rand`, equivalent to `crypto.randomUUID()`.
- **TIFF quality**: `govips`/libvips applies the quality parameter to TIFF just like Sharp does.
- The resulting binary is a native executable with no Node.js/npm dependency — it only needs
  libvips installed on the system at runtime (dynamically linked).



## 📊 Benchmark

Conversion over the same input image set, using govips (libvips) with parallel processing (`workerCount = NumCPU() / 2`).

| Format | Images processed | Total time   | Average speed |
|--------|-------------------|--------------|----------------|
| WebP   | 1409               | 18s          | ~78.3 img/s    |
| AVIF   | 210                | 1m 20s (80s) | ~2.6 img/s     |

> **Note:** AVIF is noticeably slower than WebP because its encoding algorithm (based on AV1) is computationally much more expensive,
 even with `Effort: 0` (the fastest available setting). This is an expected trade-off: AVIF typically achieves better compression at the same visual quality,
 at the cost of higher CPU time per image.

**Test environment:**
- CPU: AMD Ryzen 9 5900X (12 cores / 24 threads) @ 4.95 GHz
- RAM: 32 GiB
- Disk: NVMe SSD (btrfs)
- OS: openSUSE Tumbleweed x86_64
- Workers used: `NumCPU() / 2`



## Project structure

```
.
├── go.mod
├── go.sum
├── internal
│   ├── cliopts
│   │   └── cliopts.go       # CLI argument/flag parsing
│   ├── converter
│   │   └── converter.go     # Image conversion logic (govips/libvips)
│   ├── format
│   │   └── format.go        # Supported format definitions
│   ├── naming
│   │   └── naming.go        # File naming strategies (original, snowflake, uuid)
│   ├── runner
│   │   ├── progress.go      # Progress reporting
│   │   └── runner.go        # Orchestrates the conversion workflow/workers
│   └── scanner
│       └── scanner.go       # Directory/file scanning
├── main.go
└── README.md
```
