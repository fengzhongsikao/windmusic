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
  /** 以当前播放点为中心，左右各展示的时间窗（秒） */
  localWindowSec?: number;
  live?: boolean;
  phaseMs?: number;
  spreadMode?: WaveformSpreadMode;
}

interface SpreadContext {
  t: number;
  dist: number;
  skew: number;
  barPhase: number;
  seedS1: number;
  seedS3: number;
}

/** 海浪式起伏：多频率、每柱独立相位，无横向时间轴漂移。 */
function spreadMotionCenterOut(ctx: SpreadContext): number {
  const { t, barPhase, seedS1, seedS3 } = ctx;

  const f1 = 3.6 + seedS1 * 2.8;
  const f2 = 5.4 + seedS3 * 2.2;
  const f3 = 7.3 + seedS1 * seedS3 * 3.1;
  const f4 = 4.7 + seedS3 * 1.6;

  const waveA = Math.sin(t * f1 + barPhase);
  const waveB = Math.sin(t * f2 - barPhase * 1.41);
  const waveC = Math.sin(t * f3 + barPhase * 0.73 + seedS1 * 4.2);
  const waveD = Math.sin(t * f4 - barPhase * 0.97 + seedS3 * 3.8);
  const foam = Math.sin(t * (9.5 + seedS1 * 4.5) + barPhase * 2.15) * 0.06;

  const swell = 0.5 + waveA * 0.2 + waveB * 0.17 + waveC * 0.14 + waveD * 0.11 + foam;
  return Math.max(0.26, Math.min(1.12, swell));
}

/** 播放中的动态包络：两种模式视觉差异明显，波形本体仍居中。 */
function spreadMotion(mode: WaveformSpreadMode, ctx: SpreadContext): number {
  const { t, dist, skew, barPhase } = ctx;

  switch (mode) {
    case 'right-left': {
      const barPos = skew + 0.5;
      const front = 1 - ((t * 1.55) % 1);
      const delta = barPos - front;
      const band = Math.exp(-delta * delta * 55);
      const wake = Math.exp(-Math.pow(delta - 0.1, 2) * 38) * 0.38;
      const shimmer = 0.14 * Math.sin(t * 10.5 + barPhase);
      return 0.3 + band * 0.92 + wake + shimmer;
    }
    default:
      return spreadMotionCenterOut(ctx);
  }
}

/** 当前播放点附近采样，不把柱位映射成时间轴，避免右→左漂移。 */
function sampleOceanPeak(
  peaks: Float32Array,
  centerIdx: number,
  halfSpan: number,
  seedS1: number,
  seedS3: number,
  live: boolean,
): number {
  const micro = ((seedS1 - 0.5) * 0.06 + (seedS3 - 0.5) * 0.04) * halfSpan;
  const peak = samplePeakAt(peaks, centerIdx + micro, live);
  const texture = 0.62 + seedS1 * 0.24 + seedS3 * 0.18;
  return peak * texture;
}

/**
 * 居中对齐时间轴：中间=此刻，左右=前后片段；播放时叠加模式化动态。
 */
export function samplePeaksBarHeights(
  peaks: Float32Array,
  duration: number,
  currentTime: number,
  barCount: number,
  out: Float32Array,
  options: SamplePeaksOptions = {},
): boolean {
  const { localWindowSec = 0.55, live = false, phaseMs = 0, spreadMode = 'center-out' } = options;

  if (peaks.length === 0 || duration <= 0 || barCount <= 0) {
    return false;
  }

  const centerIdx = (currentTime / duration) * (peaks.length - 1);
  const halfSpan = Math.max(3, ((localWindowSec / duration) * peaks.length) / 2);
  let maxBar = 0;

  const barCenter = (barCount - 1) * 0.5;
  const t = phaseMs * 0.001;

  for (let i = 0; i < barCount; i++) {
    const dist = Math.abs(i - barCenter) / Math.max(1, barCenter);
    const skew = i / Math.max(1, barCount - 1) - 0.5;
    const seed = barSeed(i, barCount);
    const env = viewportEnvelope(i, barCount);
    const barPhase = seed.s2 * Math.PI * 2 + i * 0.38;

    let base: number;
    if (spreadMode === 'right-left') {
      // 线性时间轴：随播放推进，波形自右向左流过
      const offsetNorm = (i - barCenter) / Math.max(1, barCenter);
      const peakIdx = centerIdx + offsetNorm * halfSpan + (seed.s1 - 0.5) * halfSpan * 0.06;
      base = samplePeakAt(peaks, peakIdx, live);
    } else {
      // 原地海浪：峰值取自当前时刻，高度差由动画驱动
      base = sampleOceanPeak(peaks, centerIdx, halfSpan, seed.s1, seed.s3, live);
    }

    let value = base * env;

    if (live && value > 0.015) {
      const motion = spreadMotion(spreadMode, {
        t,
        dist,
        skew,
        barPhase,
        seedS1: seed.s1,
        seedS3: seed.s3,
      });
      value = Math.min(1, base * env * motion);
    }

    out[i] = value;
    if (value > maxBar) maxBar = value;
  }

  return maxBar > 0.02;
}

/** 无峰值缓存时：纯动画波形，保证模式切换可见。 */
export function sampleSpreadMotionBarHeights(
  barCount: number,
  out: Float32Array,
  options: Pick<SamplePeaksOptions, 'live' | 'phaseMs' | 'spreadMode'>,
): boolean {
  const { live = false, phaseMs = 0, spreadMode = 'center-out' } = options;

  if (!live) {
    computeRestBarHeights(barCount, out);
    return true;
  }

  const t = phaseMs * 0.001;
  const barCenter = (barCount - 1) * 0.5;
  let maxBar = 0;

  for (let i = 0; i < barCount; i++) {
    const dist = Math.abs(i - barCenter) / Math.max(1, barCenter);
    const skew = i / Math.max(1, barCount - 1) - 0.5;
    const seed = barSeed(i, barCount);
    const env = viewportEnvelope(i, barCount);
    const barPhase = seed.s2 * Math.PI * 2 + i * 0.38;
    const base = env * (0.42 + seed.s3 * 0.48);
    const motion = spreadMotion(spreadMode, {
      t,
      dist,
      skew,
      barPhase,
      seedS1: seed.s1,
      seedS3: seed.s3,
    });
    const value = Math.min(1, base * motion);
    out[i] = value;
    if (value > maxBar) maxBar = value;
  }

  return maxBar > 0.02;
}

/** 实时音频 + 扩散模式：峰值加载失败时的主路径。 */
export function sampleAnalyserSpreadBarHeights(
  analyser: AnalyserNode,
  barCount: number,
  out: Float32Array,
  options: Pick<SamplePeaksOptions, 'phaseMs' | 'spreadMode'>,
): boolean {
  const scratch = new Float32Array(barCount);
  const spreadMode = options.spreadMode ?? 'center-out';
  const phaseMs = options.phaseMs ?? 0;
  const hasSignal = sampleLiveBarHeights(analyser, barCount, scratch);

  if (!hasSignal) {
    return sampleSpreadMotionBarHeights(barCount, out, {
      live: true,
      phaseMs,
      spreadMode,
    });
  }

  const t = phaseMs * 0.001;
  const barCenter = (barCount - 1) * 0.5;
  let maxBar = 0;

  for (let i = 0; i < barCount; i++) {
    const dist = Math.abs(i - barCenter) / Math.max(1, barCenter);
    const skew = i / Math.max(1, barCount - 1) - 0.5;
    const seed = barSeed(i, barCount);
    const barPhase = seed.s2 * Math.PI * 2 + i * 0.38;
    const motion = spreadMotion(spreadMode, {
      t,
      dist,
      skew,
      barPhase,
      seedS1: seed.s1,
      seedS3: seed.s3,
    });
    const value = Math.min(1, scratch[i] * (motion / 0.62));
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

/** 居中实时波浪：当前播放点在正中，播放时高亮中心轴。 */
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
  const centerBar = (barCount - 1) * 0.5;

  for (let i = 0; i < barCount; i++) {
    const x = offsetX + i * WAVEFORM_BAR_SPACING;
    const h = heights[i];
    const distFromCenter = Math.abs(i - centerBar) / Math.max(1, centerBar);
    const centerBoost = active ? 1 + Math.max(0, 1 - distFromCenter * 2.2) * 0.12 : 1;
    const halfHeight = Math.max(WAVEFORM_BAR_MIN_HEIGHT / 2, h * maxHalfHeight * centerBoost);
    const alpha = active ? 0.5 + h * 0.5 : 0.25 + h * 0.22;
    drawBar(ctx, x, centerY, WAVEFORM_BAR_WIDTH + 1, halfHeight * 1.08, color, alpha * 0.16);
    drawBar(ctx, x, centerY, WAVEFORM_BAR_WIDTH, halfHeight, color, alpha);
  }
}
