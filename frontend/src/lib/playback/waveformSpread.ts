export type WaveformSpreadMode = 'center-out' | 'right-left';

export const DEFAULT_WAVEFORM_SPREAD: WaveformSpreadMode = 'center-out';

export const WAVEFORM_SPREAD_OPTIONS: {
  id: WaveformSpreadMode;
  label: string;
  description: string;
}[] = [
  {
    id: 'center-out',
    label: '中间向两边',
    description: '中间与两侧自然起伏，如海浪般不规则涌动',
  },
  {
    id: 'right-left',
    label: '右边向左边',
    description: '光带自右向左扫过波形',
  },
];

export function normalizeWaveformSpreadMode(value: string | undefined): WaveformSpreadMode {
  if (value === 'right-left') {
    return value;
  }
  return DEFAULT_WAVEFORM_SPREAD;
}
