import type { PlayerTrack } from '@/stores/playback/player.svelte';
import { trackPlaybackKey } from '@/stores/playback/lyrics';
import {
  coverLookupKeys,
  fetchLocalSongCovers,
  localPathFromPlayerTrack,
  resolveCoverFromMaps,
} from '@/lib/library/localMusic';
import { localLibrary, applyLocalCoverForPath } from '@/stores/library/localLibrary.svelte';
import { fetchCoverUrl } from '@/lib/wails/wailsPlayer';
import defaultCover from '@/assets/images/default.jpg';

export { defaultCover as playerDefaultCover };

/** 订阅当前曲目封面（同步 local 缓存 + 异步拉取），切歌时立即重置为默认图。 */
export function bindPlayerTrackCover(getTrack: () => PlayerTrack): {
  readonly displayedCover: string;
} {
  let displayedCover = $state(defaultCover);

  $effect(() => {
    const track = getTrack();
    trackPlaybackKey(track);

    const path = localPathFromPlayerTrack(track);
    if (path) {
      void localLibrary.coverTickByPath[path];
    }

    let cancelled = false;
    displayedCover = defaultCover;

    if (path && !localLibrary.coverByPath[path]) {
      void fetchLocalSongCovers([path]).then((batch) => {
        if (cancelled) {
          return;
        }
        for (const [filePath, key] of Object.entries(batch.paths)) {
          const cover = batch.covers[key]?.trim();
          if (cover && filePath === path) {
            applyLocalCoverForPath(filePath, cover);
          }
        }
      });
    }

    const sync = resolveCoverFromMaps(track.coverUrl, coverLookupKeys(path, track.id), [
      localLibrary.coverByPath,
    ]);
    if (sync) {
      displayedCover = sync;
      return () => {
        cancelled = true;
      };
    }

    void (async () => {
      const resolved = await fetchCoverUrl(track);
      if (cancelled) {
        return;
      }
      if (!resolved || resolved === defaultCover) {
        displayedCover = defaultCover;
        return;
      }

      const img = new Image();
      img.onload = () => {
        if (!cancelled) {
          displayedCover = resolved;
        }
      };
      img.onerror = () => {
        if (!cancelled) {
          displayedCover = defaultCover;
        }
      };
      img.src = resolved;
    })();

    return () => {
      cancelled = true;
    };
  });

  return {
    get displayedCover() {
      return displayedCover;
    },
  };
}
