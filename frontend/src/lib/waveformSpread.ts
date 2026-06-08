export type WaveformSpreadMode = 'center-out' | 'edges-in' | 'right-left';

export const DEFAULT_WAVEFORM_SPREAD: WaveformSpreadMode = 'center-out';

export const WAVEFORM_SPREAD_OPTIONS: {
  id: WaveformSpreadMode;
  label: string;
  description: string;
}[] = [
  {
    id: 'center-out',
    label: '中间向两边',
    description: '波纹从正中向两侧扩散',
  },
  {
    id: 'edges-in',
    label: '两边向中间',
    description: '波纹从两侧向正中汇聚',
  },
  {
    id: 'right-left',
    label: '右边向左边',
    description: '波纹从右侧向左侧扫过',
  },
];

export function normalizeWaveformSpreadMode(value: string | undefined): WaveformSpreadMode {
  if (value === 'edges-in' || value === 'right-left') {
    return value;
  }
  return DEFAULT_WAVEFORM_SPREAD;
}
