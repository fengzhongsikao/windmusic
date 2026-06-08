export interface TrackItem {
  id: string | number;
  /** Stable key for {#each}; defaults to id when omitted. */
  listKey?: string;
  title: string;
  artist: string;
  album: string;
  duration: string;
  /** File size label for local tracks (e.g. "8.2 MB") */
  size?: string;
  coverUrl?: string;
}
