# go-zipper

A Go-based file compression and decompression utility exposing a RESTful HTTP API. The project implements a two-stage compression pipeline combining LZ77 tokenization and Huffman encoding, packaged into a custom archive format with CRC checksum verification.

---

## Table of Contents

- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Installation and Setup](#installation-and-setup)
- [Configuration](#configuration)
- [Usage Guide](#usage-guide)
- [Features](#features)
- [Contributing](#contributing)

---

## Project Structure

```
go-zipper/
├── main.go                          # Entry point; initializes and starts the HTTP server
├── go.mod                           # Module definition and dependency declarations
├── go.sum                           # Dependency checksums
│
├── api/                             # HTTP layer: routing and request handlers
│   ├── server.go                    # Server struct, startup, and graceful shutdown logic
│   ├── router.go                    # Route definitions using the chi router
│   ├── handlers.go                  # Handler functions for compress and decompress endpoints
│   └── middleware.go                # Shared middleware (logging, recovery, etc.)
│
├── archive/                         # Archive format definition and orchestration
│   ├── archive.go                   # Archive struct, metadata schema, and format constants
│   ├── writer.go                    # Encodes and writes compressed data to archive format
│   └── reader.go                    # Reads and validates archive files for decompression
│
├── compressor/
│   ├── huffman/                     # Huffman encoding and decoding
│   │   ├── huffman.go               # Tree construction, encoding table generation
│   │   ├── encoder.go               # Bitstream encoding using the Huffman tree
│   │   └── decoder.go               # Bitstream decoding back to token stream
│   │
│   └── lz77/                        # LZ77 sliding window compression
│       ├── lz77.go                  # Core LZ77 algorithm: tokenization and back-reference resolution
│       ├── encoder.go               # Serializes LZ77 tokens to byte sequences
│       └── decoder.go               # Deserializes byte sequences back to LZ77 tokens
│
└── utils/                           # Shared utilities
    ├── fileutils.go                 # File reading, writing, and multipart form helpers
    └── checksum.go                  # CRC checksum computation and verification
```

### Package Descriptions

| Package | Responsibility |
|---|---|
| `api/` | Defines the HTTP server, registers routes via the `chi` router, and implements handlers that bridge HTTP requests to the compression logic. |
| `archive/` | Owns the custom archive format. The writer serialises compressed payloads and metadata (filenames, sizes, checksums) into a single archive file. The reader parses and validates that format before decompression. |
| `compressor/huffman/` | Builds a Huffman tree from symbol frequencies, produces an encoding table, and performs bitstream encoding and decoding. |
| `compressor/lz77/` | Implements the LZ77 sliding-window algorithm. Produces a stream of literal/back-reference tokens during compression and resolves those tokens back to raw bytes during decompression. |
| `utils/` | Provides file I/O helpers and CRC checksum routines used by both the archive and API layers to ensure data integrity. |

---

## Architecture

### Compression Pipeline

```mermaid
flowchart TD
    A([Input Files]) --> B[Multipart Form Parser\napi/handlers.go]
    B --> C[LZ77 Encode\ncompressor/lz77/encoder.go]
    C --> D[Serialize Tokens\ncompressor/lz77/lz77.go]
    D --> E[Huffman Encode\ncompressor/huffman/encoder.go]
    E --> F[CRC Checksum\nutils/checksum.go]
    F --> G[Archive Writer\narchive/writer.go]
    G --> H([.gz Archive returned to client])

    style A fill:#1e293b,color:#f1f5f9,stroke:#334155
    style H fill:#1e293b,color:#f1f5f9,stroke:#334155
```

### Decompression Pipeline

```mermaid
flowchart TD
    A([Archive File from client]) --> B[Archive Reader\narchive/reader.go]
    B --> C{Checksum Verification\nutils/checksum.go}
    C -- Valid --> D[Huffman Decode\ncompressor/huffman/decoder.go]
    C -- Invalid --> E([Error: Corrupt Archive])
    D --> F[LZ77 Decode\ncompressor/lz77/decoder.go]
    F --> G[Reconstruct Files\nutils/fileutils.go]
    G --> H([Output Files returned to client])

    style A fill:#1e293b,color:#f1f5f9,stroke:#334155
    style H fill:#1e293b,color:#f1f5f9,stroke:#334155
    style E fill:#7f1d1d,color:#fca5a5,stroke:#991b1b
```

### System Architecture Overview

```mermaid
flowchart LR
    subgraph Client
        CL([HTTP Client])
    end

    subgraph API Layer
        R[chi Router]
        H[Handlers]
        MW[Middleware]
    end

    subgraph Compression Engine
        LZ[LZ77 Codec]
        HF[Huffman Codec]
    end

    subgraph Archive Layer
        AW[Archive Writer]
        AR[Archive Reader]
        CS[CRC Checksum]
    end

    CL -->|POST multipart/form-data| R
    R --> MW --> H
    H -->|compress path| LZ --> HF --> AW
    H -->|decompress path| AR --> CS
    CS -->|pass| HF --> LZ
    AW -->|archive bytes| CL
    LZ -->|decompressed files| CL
```

### Archive Format Layout

Each archive produced by go-zipper follows this binary layout:

```
+------------------+
|  Magic Header    |  4 bytes  – format identifier
+------------------+
|  Version         |  1 byte   – archive format version
+------------------+
|  File Count      |  4 bytes  – number of contained files
+------------------+
|  [ File Entry ]  |  repeated for each file:
|    Filename Len  |  2 bytes
|    Filename      |  variable
|    Original Size |  8 bytes
|    Payload Size  |  8 bytes
|    CRC Checksum  |  4 bytes
|    Payload       |  variable – Huffman-encoded LZ77 token stream
+------------------+
```

---

## Installation and Setup

### Prerequisites

| Requirement | Version |
|---|---|
| Go | 1.21 or later |
| chi router | v5.x |

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/your-org/go-zipper.git
cd go-zipper

# Download dependencies
go mod download

# Build the binary
go build -o go-zipper ./...

# Run the binary
./go-zipper
```

### Run Directly with Go

```bash
go run main.go
```

### Run Tests

```bash
go test ./...
```

---

## Configuration

The server is configured through environment variables and has sensible defaults when none are provided.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | TCP port the HTTP server listens on |
| `HOST` | `0.0.0.0` | Network interface to bind |

### Example Environment Configuration

```bash
export PORT=9090
./go-zipper
```

### Server Options (Programmatic)

The `api.Server` struct accepts the following options:

```go
type Server struct {
    Host    string  // Network interface (default: "0.0.0.0")
    Port    string  // Port number     (default: "8080")
    Version string  // API version     (default: "v1")
}
```

### Graceful Shutdown

The server listens for `SIGINT` and `SIGTERM` signals. Upon receiving either signal, it stops accepting new connections and waits for in-flight requests to complete before exiting. The shutdown timeout is configurable within `api/server.go`.

```
Signal received (SIGINT / SIGTERM)
        |
        v
Stop accepting new connections
        |
        v
Wait for active requests (timeout: 30s)
        |
        v
Release resources and exit cleanly
```

---

## Usage Guide

### API Endpoints

#### `GET /health`

Returns the health status of the server. Use this endpoint to verify the service is running and responsive.

**Request**

```bash
curl -X GET http://localhost:8080/health
```

**Response**

```json
{
  "status": "ok"
}
```

---

#### `POST /compress`

Accepts one or more files via a `multipart/form-data` request and returns a single compressed archive.

**Constraints**

- Maximum request body size: 32 MB
- Content-Type must be `multipart/form-data`
- Files must be supplied under the form field name `files`

**Request**

```bash
curl -X POST http://localhost:8080/compress \
  -F "files=@document.txt" \
  -F "files=@image.png" \
  --output archive.fzp
```

**Response**

On success, the server responds with HTTP `200` and streams the archive binary as the response body.

| Header | Value |
|---|---|
| `Content-Type` | `application/octet-stream` |
| `Content-Disposition` | `attachment; filename="archive.fzp"` |
| `Content-Length` | `integer` |

**Error Responses**

| Status | Reason |
|---|---|
| `400 Bad Request` | Missing or malformed multipart form data |
| `413 Request Entity Too Large` | Upload exceeds the 32 MB limit |
| `500 Internal Server Error` | Compression or archive writing failure |

---

#### `POST /decompress`

Accepts a previously created go-zipper archive and returns the decompressed files. If the archive contains multiple files, the response is a ZIP bundle of the reconstructed files.

**Request**

```bash
curl -X POST http://localhost:8080/decompress \
  -F "file=@archive.fzp" \
  --output decompressed.fzp.
```

**Response**

On success, the server responds with HTTP `200` and streams the decompressed content.

| Header | Value |
|---|---|
| `Content-Type` | `application/octet-stream` |
| `Content-Disposition` | `attachment; filename="decompressed.fzp"` |
| `Content-Length` | `integer` |

**Error Responses**

| Status | Reason |
|---|---|
| `400 Bad Request` | No archive file provided, or file is not a valid go-zipper archive |
| `422 Unprocessable Entity` | CRC checksum mismatch; archive is corrupt |
| `500 Internal Server Error` | Decompression or file reconstruction failure |

---

### Request and Response Format Summary

```
POST /compress
  Content-Type: multipart/form-data
  Body:         files[] -> binary file uploads
  Response:     application/octet-stream (archive binary)

POST /decompress
  Content-Type: multipart/form-data
  Body:         file   -> single archive binary
  Response:     application/octet-stream (decompressed content)

GET /health
  Response:     application/json
```

---

## Features

### Two-Stage Compression

go-zipper applies compression in two sequential stages:

1. **LZ77 Encoding** — A sliding-window algorithm scans the input and replaces repeated byte sequences with back-references `(offset, length)`. This reduces structural redundancy in the data stream before any entropy coding is applied.

2. **Huffman Encoding** — The token stream produced by LZ77 is analysed for symbol frequency. A Huffman tree is constructed and used to assign shorter bit codes to more frequent symbols, yielding further size reduction through optimal entropy coding.

This combination is effective on text-heavy files and structured binary data. The achievable compression ratio depends on the entropy and repetition characteristics of the input.

### CRC Checksum Verification

Each file entry in the archive stores a CRC-32 checksum computed from the original uncompressed data. During decompression, the checksum is recomputed and compared against the stored value. Any mismatch causes the operation to abort with an error, protecting against silent data corruption caused by transmission errors or storage faults.

### Custom Archive Format

The archive format is self-describing: it embeds filenames, original sizes, compressed payload sizes, and checksums for each file entry. This allows the decompressor to validate and reconstruct files independently without relying on external metadata or sidecar files.

### Multiple File Support

A single compress request may include multiple files. All files are encoded and stored within one archive. During decompression, all original files are reconstructed from that single archive in their original form.

### RESTful HTTP API

All functionality is exposed over HTTP using standard request and response conventions. The server requires no client-side libraries; any HTTP client capable of sending multipart form data — including curl, Postman, or custom application code — can interact with the API.

### Graceful Shutdown

The server handles OS termination signals cleanly, completing any active requests before releasing resources and exiting. This makes go-zipper suitable for deployment in container-orchestrated environments where rolling restarts and zero-downtime deployments are expected.

---