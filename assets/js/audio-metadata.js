const CODEC_MAP = {
  "audio/mpeg": "MP3",
  "audio/wav": "PCM",
  "audio/wave": "PCM",
  "audio/x-wav": "PCM",
  "audio/mp4": "AAC",
  "audio/x-m4a": "AAC",
  "audio/ogg": "Vorbis",
  "audio/flac": "FLAC",
  "audio/x-flac": "FLAC",
};

const FORMAT_MAP = {
  "audio/mpeg": "mp3",
  "audio/wav": "wav",
  "audio/wave": "wav",
  "audio/x-wav": "wav",
  "audio/mp4": "m4a",
  "audio/x-m4a": "m4a",
  "audio/ogg": "ogg",
  "audio/flac": "flac",
  "audio/x-flac": "flac",
};

function getCodec(mimeType) {
  if (CODEC_MAP[mimeType]) {
    return CODEC_MAP[mimeType];
  }
  return "Unknown";
}

function getFormat(mimeType) {
  if (FORMAT_MAP[mimeType]) {
    return FORMAT_MAP[mimeType];
  }
  return "unknown";
}

function formatDuration(sec) {
  const hours = Math.floor(sec / 3600);
  const minutes = Math.floor((sec % 3600) / 60);
  const seconds = sec % 60;
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  }
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function formatFileSize(bytes) {
  if (bytes < 1024) return `${bytes} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let i = -1;
  do {
    bytes /= 1024;
    i++;
  } while (bytes >= 1024 && i < units.length - 1);
  return `${bytes.toFixed(2)} ${units[i]}`;
}

function setHiddenInput(id, value) {
  const input = document.getElementById(id);
  if (input) {
    input.value = value;
  }
}

// main extraction function
window.extractAudioMetadata = function (el) {
  const file = el.files[0];
  if (!file) return;

  const fileSize = file.size;
  const mimeType = file.type;
  const format = getFormat(mimeType);

  // set hidden inputs for form submission
  setHiddenInput("audio_format", format);
  setHiddenInput("audio_mimeType", mimeType);
  setHiddenInput("audio_file_size", fileSize);

  // decode audio to get duration and sample rate
  const audioContext = new (window.AudioContext || window.webkitAudioContext)();
  const reader = new FileReader();
  reader.onload = function (event) {
    const arrayBuffer = event.target.result;
    audioContext.decodeAudioData(
      arrayBuffer,
      function (decodedData) {
        const duration = decodedData.duration;
        const sampleRate = decodedData.sampleRate;
        const channelCount = decodedData.numberOfChannels;
        const bitrate = Math.round((fileSize * 8) / duration / 1000);

        // set hidden inputs for form submission
        setHiddenInput("audio_duration", duration);
        setHiddenInput("audio_sample_rate", sampleRate);
        setHiddenInput("audio_channel_count", channelCount);
        setHiddenInput("audio_bitrate", bitrate);

        // update datastar signal
        const detail = {
          only: {
            audio_duration: formatDuration(duration),
            audio_sample_rate: sampleRate.toLocaleString() + " Hz",
            audio_channel_count:
              channelCount === 1
                ? "Mono"
                : channelCount === 2
                  ? "Stereo"
                  : channelCount + " channels",
            audio_bitrate: bitrate + " kbps",
            audio_format: format.toUpperCase(),
            audio_mime_type: mimeType,
            audio_file_size: formatFileSize(fileSize),
            metadata_extracted: "true",
          },
        };

        document.dispatchEvent(new CustomEvent('datastar-patch-signals', { detail }));

      audioContext.close();
      },
      function (error) {
        console.error("Error decoding audio data:", error);
      },
    );
  };
  reader.readAsArrayBuffer(file);
};
