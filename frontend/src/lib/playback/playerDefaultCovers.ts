import defaultCover from '@/assets/images/default.jpg';
import localDefaultCover from '@/assets/images/pai.png';
import { localPathFromPlayerTrack } from '@/lib/library/localMusic';

export { defaultCover, localDefaultCover };

export function resolvePlayerDefaultCover(track: Parameters<typeof localPathFromPlayerTrack>[0]): string {
  return localPathFromPlayerTrack(track) ? localDefaultCover : defaultCover;
}
