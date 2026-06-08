import type { WaveformSpreadMode } from '@/lib/waveformSpread';

export const WAVEFORM_BAR_WIDTH = 2;
export const WAVEFORM_BAR_GAP = 2;
export const WAVEFORM_BAR_SPACING = WAVEFORM_BAR_WIDTH + WAVEFORM_BAR_GAP;
export const WAVEFORM_BAR_RADIUS = 1;
export const WAVEFORM_BAR_MIN_HEIGHT = 2;

export interface WaveformColors {
  wave: string;
  progress: string;
}

export interface DrawLiveWaveformOptions {
  ctx: CanvasRenderingContext2D;
  heights: Float32Array;
  viewportWidth: number;
  height: number;
  colors: WaveformColors;
  active: boolean;
  devicePixelRatio?: number;
}

export function computeVisibleBarCount(viewportWidth: number): number {
  return Math.max(36, Math.floor(viewportWidth / WAVEFORM_BAR_SPACING));
}

export function viewportEnvelope(index: number, count: number): number {
  const t = index / Math.max(1, count - 1);
  const dist = Math.abs(t - 0.5) / 0.5;
  return Math.max(0.1, 1 - dist * dist * 0.9);
}

const PHI = 0.6180339887;

interface BarSeed {
  s1: number;
  s2: number;
  s3: number;
}

function barSeed(index: number, count: number): BarSeed {
  const base = (index + 1) / (count + 1);
  return {
    s1: (base * 1.113) % 1,
    s2: (base * PHI * 7.31) % 1,
    s3: (base * PHI * PHI * 13.7) % 1,
  };
}

function scatterIndex(length: number, seed: number, salt: number): number {
  if (length <= 0) return 0;
  return Math.floor(((seed * 997 + salt * 0.137) % 1) * length) % length;
}

export function measureAudioLevel(analyser: AnalyserNode): { rms: number; peak: number } {
  const timeData = new Uint8Array(analyser.fftSize);
  analyser.getByteTimeDomainData(timeData);
  let sumSq = 0;
  let peak = 0;

  for (let i = 0; i < timeData.length; i++) {
    const v = Math.abs(timeData[i] - 128) / 128;
    sumSq += v * v;
    if (v > peak) peak = v;
  }

  return { rms: Math.sqrt(sumSq / timeData.length), peak };
}

export function isAudioAudible(analyser: AnalyserNode): boolean {
  const { rms, peak } = measureAudioLevel(analyser);
  return rms > 0.012 || peak > 0.045;
}

/** 仅根据真实音频采样绘制，无信号时返回 false。 */
export function sampleLiveBarHeights(
  analyser: AnalyserNode,
  barCount: number,
  out: Float32Array,
): boolean {
  const timeData = new Uint8Array(analyser.fftSize);
  const freqData = new Uint8Array(analyser.frequencyBinCount);
  analyser.getByteTimeDomainData(timeData);
  analyser.getByteFrequencyData(freqData);

  const { rms, peak } = measureAudioLevel(analyser);
  if (!isAudioAudible(analyser)) {
    for (let i = 0; i < barCount; i++) {
      out[i] = 0;
    }
    return false;
  }

  const usableFreq = Math.max(8, Math.floor(freqData.length * 0.62));
  const gain = Math.min(1.15, rms * 3.8 + peak * 0.55);
  let maxBar = 0;

  for (let i = 0; i < barCount; i++) {
    const env = viewportEnvelope(i, barCount);
    const seed = barSeed(i, barCount);

    const t1 = scatterIndex(timeData.length, seed.s1, 0.21);
    const t2 = scatterIndex(timeData.length, seed.s2, 0.57);
    const t3 = scatterIndex(timeData.length, seed.s3, 0.83);
    const wave =
      (Math.abs(timeData[t1] - 128) +
        Math.abs(timeData[t2] - 128) +
        Math.abs(timeData[t3] - 128)) /
      (128 * 3);

    const f1 = scatterIndex(usableFreq, seed.s1, 0.39);
    const f2 = scatterIndex(usableFreq, seed.s3, 0.71);
    const freq = (freqData[f1] + freqData[f2]) / (255 * 2);

    const value = Math.min(1, env * gain * (wave * 0.68 + freq * 0.42));
    out[i] = value;
    if (value > maxBar) maxBar = value;
  }

  return maxBar > 0.08;
}

export function computeRestBarHeights(barCount: number, out: Float32Array): void {
  for (let i = 0; i < barCount; i++) {
    out[i] = viewportEnvelope(i, barCount) * 0.1;
  }
}

function samplePeakAt(peaks: Float32Array, index: number, sharp = false): number {
  const clamped = Math.min(peaks.length - 1, Math.max(0, index));
  if (sharp) {
    return peaks[Math.round(clamped)];
  }
  const i = Math.floor(clamped);
  const frac = clamped - i;
  const next = Math.min(peaks.length - 1, i + 1);
  return peaks[i] * (1 - frac) + peaks[next] * frac;
}

export interface SamplePeaksOptions {
  /** 围绕此刻播放点采样的时间窗（秒），越小越像原地起伏的海浪 */
  localWindowSec?: number;
  /** 播放中：叠加轻微多频起伏（仍由真实峰值驱动） */
  live?: boolean;
  phaseMs?: number;
  spreadMode?: WaveformSpreadMode;
}

function spreadRipple(
  mode: WaveformSpreadMode,
  t: number,
  dist: number,
  pos: number,
  barIndex: number,
  bandJitter: number,
  seedS2: number,
  seedS3: number,
): number {
  const barPhase = barIndex * 0.42 + seedS2 * Math.PI * 2;

  switch (mode) {
    case 'edges-in':
      return (
        Math.sin(t * 6.8 + dist * 10.5 + barPhase + bandJitter) * 0.24 +
        Math.sin(t * 5.1 + dist * 7.2 + seedS3 * 1.4) * 0.16 +
        (1 - dist) * Math.sin(t * 4.6 + bandJitter) * 0.12
      );
    case 'right-left':
      return (
        Math.sin(t * 6.4 - pos * 11.0 + barPhase + bandJitter) * 0.26 +
        Math.sin(t * 4.9 - pos * 7.5 + seedS3 * 1.2) * 0.15
      );
    default:
      return (
        Math.sin(t * 7.0 - dist * 10.8 + barPhase + bandJitter) * 0.25 +
        Math.sin(t * 5.3 - dist * 7.0 + seedS3 * 1.3) * 0.17 +
        (1 - dist) * Math.sin(t * 4.8 + bandJitter) * 0.11
      );
  }
}

/**
 * 居中海浪式采样：峰值以播放点为中心对称读取，
 * 涟漪相位按距中心距离扩散，从中间向两侧走。
 */
export function samplePeaksBarHeights(
  peaks: Float32Array,
  duration: number,
  currentTime: number,
  barCount: number,
  out: Float32Array,
  options: SamplePeaksOptions = {},
): boolean {
  const { localWindowSec = 0.42, live = false, phaseMs = 0, spreadMode = 'center-out' } = options;

  if (peaks.length === 0 || duration <= 0 || barCount <= 0) {
    return false;
  }

  const centerIdx = (currentTime / duration) * (peaks.length - 1);
  const halfSpan = Math.max(2, ((localWindowSec / duration) * peaks.length) / 2);
  const centerEnergy = samplePeakAt(peaks, centerIdx);
  let maxBar = 0;

  const barCenter = (barCount - 1) * 0.5;
  const peakSharp = live;

  for (let i = 0; i < barCount; i++) {
    const env = viewportEnvelope(i, barCount);
    const dist = Math.abs(i - barCenter) / Math.max(1, barCenter);
    const pos = i / Math.max(1, barCount - 1);
    const seed = barSeed(i, barCount);

    const radial = halfSpan * (0.28 + (spreadMode === 'right-left' ? pos : dist) * 0.72);
    const micro =
      spreadMode === 'right-left'
        ? (seed.s1 - 0.5) * halfSpan * 0.18
        : (seed.s1 - 0.5) * halfSpan * 0.1;
    const anchor = centerIdx + micro;
    const centerPeak = samplePeakAt(peaks, anchor, peakSharp);
    const forwardPeak = samplePeakAt(
      peaks,
      anchor + (spreadMode === 'right-left' ? radial * 0.65 : radial),
      peakSharp,
    );
    const backPeak = samplePeakAt(
      peaks,
      anchor - (spreadMode === 'right-left' ? radial * 0.35 : radial * 0.85),
      peakSharp,
    );
    const wave = live
      ? Math.max(centerPeak, forwardPeak * 0.82, backPeak * 0.72)
      : (centerPeak + forwardPeak + backPeak) / 3;

    let value = Math.min(1, env * wave);

    if (live && value > 0.04 && centerEnergy > 0.03) {
      const t = phaseMs * 0.001;
      const bandJitter = seed.s2 * Math.PI * 2;
      const ripple = spreadRipple(
        spreadMode,
        t,
        dist,
        pos,
        i,
        bandJitter,
        seed.s2,
        seed.s3,
      );
      const rippleLift = Math.max(0, ripple) * 0.55;
      const rippleDip = Math.min(0, ripple) * 0.35;
      value = Math.min(1, value * (1 + rippleDip) + env * rippleLift);
    }

    out[i] = value;
    if (value > maxBar) maxBar = value;
  }

  return maxBar > 0.02;
}

export function smoothHeights(
  current: Float32Array,
  target: Float32Array,
  factor: number,
): void {
  const count = Math.min(current.length, target.length);
  for (let i = 0; i < count; i++) {
    current[i] += (target[i] - current[i]) * factor;
  }
}

/** 帧率无关的指数平滑：上升略快、下降略慢，更接近海浪收放。 */
export function smoothHeightsDt(
  current: Float32Array,
  target: Float32Array,
  dtSeconds: number,
  attackHz = 10,
  releaseHz = 6,
): void {
  const count = Math.min(current.length, target.length);
  const dt = Math.max(0, Math.min(0.05, dtSeconds));

  for (let i = 0; i < count; i++) {
    const diff = target[i] - current[i];
    const rate = diff > 0 ? attackHz : releaseHz;
    const factor = 1 - Math.exp(-rate * dt);
    current[i] += diff * factor;
  }
}

function drawBar(
  ctx: CanvasRenderingContext2D,
  x: number,
  centerY: number,
  width: number,
  halfHeight: number,
  color: string,
  alpha: number,
) {
  const top = centerY - halfHeight;
  const barHeight = halfHeight * 2;
  ctx.globalAlpha = alpha;
  ctx.fillStyle = color;
  ctx.beginPath();
  if (WAVEFORM_BAR_RADIUS > 0 && 'roundRect' in ctx) {
    ctx.roundRect(x, top, width, barHeight, WAVEFORM_BAR_RADIUS);
  } else {
    ctx.rect(x, top, width, barHeight);
  }
  ctx.fill();
  ctx.globalAlpha = 1;
}

/** 居中实时波浪：仅展示此刻播放片段，不画未来时间轴。 */
export function drawLiveWaveform({
  ctx,
  heights,
  viewportWidth,
  height,
  colors,
  active,
  devicePixelRatio = 1,
}: DrawLiveWaveformOptions): void {
  const dpr = Math.max(1, devicePixelRatio);
  const pixelWidth = Math.max(1, Math.floor(viewportWidth * dpr));
  const pixelHeight = Math.max(1, Math.floor(height * dpr));

  if (ctx.canvas.width !== pixelWidth || ctx.canvas.height !== pixelHeight) {
    ctx.canvas.width = pixelWidth;
    ctx.canvas.height = pixelHeight;
  }

  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, viewportWidth, height);

  const barCount = heights.length;
  if (barCount === 0) {
    return;
  }

  const layoutWidth = barCount * WAVEFORM_BAR_SPACING;
  const offsetX = (viewportWidth - layoutWidth) / 2;
  const centerY = height / 2;
  const maxHalfHeight = Math.max(5, height / 2 - 3);
  const color = active ? colors.progress : colors.wave;

  for (let i = 0; i < barCount; i++) {
    const x = offsetX + i * WAVEFORM_BAR_SPACING;
    const halfHeight = Math.max(WAVEFORM_BAR_MIN_HEIGHT / 2, heights[i] * maxHalfHeight);
    const alpha = active ? 0.55 + heights[i] * 0.45 : 0.28 + heights[i] * 0.2;
    drawBar(ctx, x, centerY, WAVEFORM_BAR_WIDTH + 1, halfHeight * 1.06, color, alpha * 0.14);
    drawBar(ctx, x, centerY, WAVEFORM_BAR_WIDTH, halfHeight, color, alpha);
  }
}
