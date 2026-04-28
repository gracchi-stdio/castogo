# Audio Processing Pipeline Plan

## Goal
Server-side audio processing to normalize uploaded episode audio for podcast publication (Apple/Spotify/Google compatible).

## Deployment Target
- Bunny Magic Containers (Docker)
- FFmpeg in same container as Go app
- Bunny Storage for processed audio files
- Thread limit: `-threads 4` (audio encoding doesn't benefit from more; respects Bunny's 8-CPU pod allocation)

## Processing Steps

### 1. Upload
- Accept audio file uploads (mp3, wav, m4a, ogg, flac)
- Validate: file size limit, format check via file header (not just extension)
- Store raw upload to temporary location

### 2. FFmpeg Processing Pipeline
```
ffmpeg -i input.ext \
  -af "loudnorm=I=-16:TP=-1:LRA=11" \
  -c:a libmp3lame \
  -b:a 192k \
  -ar 44100 \
  -threads 4 \
  -y output.mp3
```

- **Loudness normalization**: EBU R128 target -16 LUFS (podcast standard), true peak -1 dBTP, loudness range 11 LU
- **Codec**: MP3 (universal compatibility across all podcast directories)
- **Bitrate**: 192kbps (good quality/size balance for speech)
- **Sample rate**: 44100 Hz (standard)

### 3. Metadata Tagging
- Write ID3 tags using `github.com/bogem/id3v2`
- Tags: title, artist, album (podcast name), track number, year, genre ("Podcast"), cover art

### 4. Storage
- Upload processed file to Bunny Storage
- Store file URL/metadata in PostgreSQL via existing episode record

### 5. Cleanup
- Delete temporary files after successful upload to Bunny Storage

## Architecture

```
internal/service/audio_processor.go   — AudioProcessor interface + FFmpeg implementation
internal/service/audio_service.go     — Orchestrates: validate → process → tag → upload → cleanup
```

### AudioProcessor Interface
```go
type AudioProcessor interface {
    Process(inputPath string, opts ProcessingOptions) (*ProcessingResult, error)
}

type ProcessingOptions struct {
    TargetLUFS    float64 // -16 for podcasts
    TargetBitrate string  // "192k"
    TargetSample  int     // 44100
    Threads       int     // 4 (Bunny-safe)
}

type ProcessingResult struct {
    OutputPath   string
    Duration     float64
    FileSize     int64
    Bitrate      int
    SampleRate   int
    Channels     int
}
```

### AudioService (orchestrator)
```go
type AudioService interface {
    ProcessAndStore(episodeID uuid.UUID, inputPath string) error
}
```

## Dependencies
- **FFmpeg** — must be installed in Docker image (`apt install ffmpeg`)
- **`github.com/bogem/id3v2`** — ID3 tag writing (pure Go)
- **`os/exec`** — FFmpeg invocation from Go

## Dockerfile Addition
```dockerfile
RUN apt-get update && apt-get install -y ffmpeg && rm -rf /var/lib/apt/lists/*
```

## Bunny-Specific Considerations
- Magic Containers get 8 CPUs per pod — limit FFmpeg to `-threads 4`
- For audio encoding (libmp3lame), this is more than sufficient
- If scaling to multiple concurrent uploads, consider a processing queue to avoid CPU saturation on a single pod

## Future Enhancements (out of scope for now)
- Two-pass loudness normalization (first pass analyzes, second pass applies)
- Waveform generation for episode pages
- Chapter markers from silence detection
- Multiple quality tiers (e.g., 128k vs 192k)
- Client-side pre-validation (duration, format check before upload)
