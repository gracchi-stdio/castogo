// Audio metadata extraction — called from episode_new_page.templ via
// `data-on:change="window.extractAudioMetadata(el)"`. Decodes the file with
// AudioContext, then dispatches an `audiometadata` CustomEvent on the input
// with computed signals (Datastar merges them into the form's signals).

interface AudioMetadataDetail {
  meta_duration: number;
  meta_sample_rate: number;
  meta_channel_count: number;
  meta_bitrate: number;
  meta_file_size: number;
  meta_format: string;
  meta_mime_type: string;
  audio_duration: string;
  audio_sample_rate: string;
  audio_channel_count: string;
  audio_bitrate: string;
  audio_format: string;
  audio_mime_type: string;
  audio_file_size: string;
}

const FORMAT_MAP: Record<string, string> = {
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

function getFormat(mimeType: string): string {
  return FORMAT_MAP[mimeType] ?? "unknown";
}

function formatDuration(sec: number): string {
  const hours = Math.floor(sec / 3600);
  const minutes = Math.floor((sec % 3600) / 60);
  const seconds = sec % 60;
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  }
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let value = bytes;
  let i = -1;
  do {
    value /= 1024;
    i++;
  } while (value >= 1024 && i < units.length - 1);
  return `${value.toFixed(2)} ${units[i]}`;
}

window.extractAudioMetadata = (el: HTMLInputElement): void => {
  const file = el.files?.[0];
  if (!file) return;

  const fileSize = file.size;
  const mimeType = file.type;
  const format = getFormat(mimeType);

  const Ctx = window.AudioContext ?? window.webkitAudioContext;
  if (!Ctx) return;
  const audioContext = new Ctx();

  const reader = new FileReader();
  reader.onload = (event) => {
    const arrayBuffer = event.target?.result;
    if (!(arrayBuffer instanceof ArrayBuffer)) return;

    audioContext.decodeAudioData(
      arrayBuffer,
      (decoded) => {
        const duration = decoded.duration;
        const sampleRate = decoded.sampleRate;
        const channelCount = decoded.numberOfChannels;
        const bitrate = Math.round((fileSize * 8) / duration / 1000);

        const detail: AudioMetadataDetail = {
          meta_duration: duration,
          meta_sample_rate: sampleRate,
          meta_channel_count: channelCount,
          meta_bitrate: bitrate,
          meta_file_size: fileSize,
          meta_format: format,
          meta_mime_type: mimeType,
          audio_duration: formatDuration(duration),
          audio_sample_rate: `${sampleRate.toLocaleString()} Hz`,
          audio_channel_count:
            channelCount === 1
              ? "Mono"
              : channelCount === 2
                ? "Stereo"
                : `${channelCount} channels`,
          audio_bitrate: `${bitrate} kbps`,
          audio_format: format.toUpperCase(),
          audio_mime_type: mimeType,
          audio_file_size: formatFileSize(fileSize),
        };

        el.dispatchEvent(new CustomEvent<AudioMetadataDetail>("audiometadata", { detail }));
        audioContext.close();
      },
      (error) => {
        console.error("Error decoding audio data:", error);
      },
    );
  };
  reader.readAsArrayBuffer(file);
};
