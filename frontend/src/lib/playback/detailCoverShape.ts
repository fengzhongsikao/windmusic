export type DetailCoverShape = 'round' | 'square';

export const DEFAULT_DETAIL_COVER_SHAPE: DetailCoverShape = 'round';

export const DETAIL_COVER_SHAPE_OPTIONS: {
  id: DetailCoverShape;
  label: string;
  description: string;
}[] = [
  {
    id: 'round',
    label: '圆形封面',
    description: '黑胶唱片风格，可开启旋转',
  },
  {
    id: 'square',
    label: '方形封面',
    description: '标准专辑封面，保持静止',
  },
];

export function normalizeDetailCoverShape(value: string | undefined): DetailCoverShape {
  if (value === 'square') {
    return value;
  }
  return DEFAULT_DETAIL_COVER_SHAPE;
}
