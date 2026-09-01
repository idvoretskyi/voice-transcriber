# Voice Transcriber

Single-binary media-to-text transcription with automatic language detection, powered by Google Gemini via Vertex AI.

[![CI](https://github.com/idvoretskyi/voice-transcriber/actions/workflows/ci.yml/badge.svg)](https://github.com/idvoretskyi/voice-transcriber/actions/workflows/ci.yml)
[![CodeQL](https://github.com/idvoretskyi/voice-transcriber/actions/workflows/codeql.yml/badge.svg)](https://github.com/idvoretskyi/voice-transcriber/actions/workflows/codeql.yml)
[![Security](https://github.com/idvoretskyi/voice-transcriber/actions/workflows/security.yml/badge.svg)](https://github.com/idvoretskyi/voice-transcriber/actions/workflows/security.yml)
[![Go version](https://img.shields.io/badge/go-1.27-00ADD8?logo=go)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Features

- **Automatic language detection** — Gemini identifies the spoken language from audio (default)
- Specify language explicitly with `--language` using an ISO 639-1 code (e.g. `uk`, `en`, `de`)
- Accepts **audio and video files** as input
- **No Cloud Storage required** — audio bytes sent inline to Gemini
- FFmpeg used only for video-to-audio extraction; audio files go straight to Gemini
- Handles files up to ~8.4 hours in a single request (no chunking)
- Selectable Gemini model via `--model` flag (default: `gemini-3.1-flash-lite-preview`)
- Single static binary — no extra runtime dependencies beyond FFmpeg for video

## Quick Start

### Prerequisites

```bash
# Go 1.27+
brew install go            # macOS
# sudo apt install golang-go  # Ubuntu/Debian

# FFmpeg (only required for video files)
brew install ffmpeg        # macOS
# sudo apt install ffmpeg    # Ubuntu/Debian
```

### Install

```bash
go install github.com/idvoretskyi/voice-transcriber/cmd/voice-transcriber@latest
```

This installs the `voice-transcriber` binary into `$(go env GOPATH)/bin`.

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Google Cloud setup

```bash
# Authenticate
gcloud auth login
gcloud auth application-default login

# Set project and enable Vertex AI
gcloud config set project YOUR_PROJECT_ID
gcloud services enable aiplatform.googleapis.com
```

The project is also read from the `GOOGLE_CLOUD_PROJECT` environment variable if set.

### Usage

```bash
# Transcribe a video file (language auto-detected, audio extracted via FFmpeg)
voice-transcriber transcribe input/meeting.mp4

# Transcribe an audio file directly
voice-transcriber transcribe input/interview.mp3

# Specify output file
voice-transcriber transcribe input/meeting.mp4 -o transcript.txt

# Force a specific language (ISO 639-1 code)
voice-transcriber transcribe input/meeting.mp4 --language uk

# Use a different model or location
voice-transcriber transcribe input/meeting.mp4 --model gemini-3-flash-preview
voice-transcriber transcribe input/meeting.mp4 --model gemini-2.5-flash --location us-central1

# Show version
voice-transcriber version
```

## CLI Reference

```
Usage:
  voice-transcriber transcribe [media-file] [flags]
  voice-transcriber version

Flags:
  --language string   Language for transcription: 'auto' for automatic detection,
                      or ISO 639-1 code (e.g. uk, en, de) (default: auto)
  --model string      Gemini model to use
                      (default: gemini-3.1-flash-lite-preview)
  --location string   Vertex AI location; Gemini 3.x models require global
                      (default: global)
  -o, --output string Output file path
                      (default: output/<name>/<name>.txt)
  -v, --verbose       Enable verbose output
  -q, --quiet         Suppress all output except results
```

## Supported Formats

| Type | Extensions |
|------|------------|
| **Audio** — sent directly to Gemini | `.wav` `.mp3` `.flac` `.ogg` `.m4a` `.aac` `.webm` `.pcm` |
| **Video** — audio extracted via FFmpeg | `.mp4` `.mkv` `.mov` `.avi` `.wmv` `.flv` `.ts` `.mpeg` `.3gp` |

Extension matching is case-insensitive. Maximum file size: 10 GB.

## Building from Source

```bash
git clone https://github.com/idvoretskyi/voice-transcriber.git
cd voice-transcriber
make build   # produces ./voice-transcriber
make test    # run tests with race detector
make lint    # run golangci-lint
```

## License

MIT — see [LICENSE](LICENSE) for details.
