const peakCache = new Map<string, Float32Array>();
const inflight = new Map<string, Promise<Float32Array | null>>();

const PEAK_COUNT = 4096;

function audioContextCtor(): typeof AudioContext | null {
  return (
    window.AudioContext ||
    (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext ||
    null
  );
}

function extractPeaks(buffer: AudioBuffer, targetPeaks: number): Float32Array {
  const channel = buffer.getChannelData(0);
  const peaks = new Float32Array(targetPeaks);
  const blockSize = Math.max(1, Math.floor(channel.length / targetPeaks));
  let globalMax = 0;

  for (let i = 0; i < targetPeaks; i++) {
    const start = i * blockSize;
    const end = Math.min(channel.length, start + blockSize);
    let peak = 0;
    for (let j = start; j < end; j++) {
      const v = Math.abs(channel[j]);
      if (v > peak) peak = v;
    }
    peaks[i] = peak;
    if (peak > globalMax) globalMax = peak;
  }

  if (globalMax > 0) {
    for (let i = 0; i < targetPeaks; i++) {
      peaks[i] /= globalMax;
    }
  }

  return peaks;
}

/** 从音频 URL 解码波形峰值（不接管 <audio> 播放，Wails 下可用）。 */
export async function loadWaveformPeaks(audioUrl: string): Promise<Float32Array | null> {
  const url = audioUrl.trim();
  if (!url) {
    return null;
  }

  const cached = peakCache.get(url);
  if (cached) {
    return cached;
  }

  const pending = inflight.get(url);
  if (pending) {
    return pending;
  }

  const request = (async () => {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        return null;
      }

      const arrayBuffer = await response.arrayBuffer();
      const Ctor = audioContextCtor();
      if (!Ctor) {
        return null;
      }

      const ctx = new Ctor();
      try {
        const decoded = await ctx.decodeAudioData(arrayBuffer.slice(0));
        const peaks = extractPeaks(decoded, PEAK_COUNT);
        peakCache.set(url, peaks);
        return peaks;
      } finally {
        void ctx.close();
      }
    } catch {
      return null;
    } finally {
      inflight.delete(url);
    }
  })();

  inflight.set(url, request);
  return request;
}

export function clearWaveformPeakCache() {
  peakCache.clear();
  inflight.clear();
}
