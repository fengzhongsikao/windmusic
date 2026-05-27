<script lang="ts">
  import { Button, ButtonGroup, Heading } from 'flowbite-svelte';
  import TrackList, { type TrackItem } from '@/components/TrackList.svelte';

  let activeTab = $state('all');

  const tabs = [
    { id: 'all', label: '全部' },
    { id: 'chinese', label: '华语' },
    { id: 'pop', label: '流行' },
    { id: 'rock', label: '摇滚' },
    { id: 'electronic', label: '电子' },
  ];

  interface Song {
    id: number;
    title: string;
    artist: string;
    album: string;
    duration: string;
  }

  const songs: Song[] = [
    { id: 1, title: '晴天', artist: '周杰伦', album: '叶惠美', duration: '4:29' },
    { id: 2, title: '稻香', artist: '周杰伦', album: '魔杰座', duration: '3:43' },
    { id: 3, title: '夜曲', artist: '周杰伦', album: '十一月的萧邦', duration: '4:34' },
    { id: 4, title: '起风了', artist: '买辣椒也用券', album: '起风了', duration: '5:12' },
    { id: 5, title: '光年之外', artist: '邓紫棋', album: '光年之外', duration: '3:52' },
    { id: 6, title: '平凡之路', artist: '朴树', album: '猎户星座', duration: '4:46' },
    { id: 7, title: '孤勇者', artist: '陈奕迅', album: '孤勇者', duration: '4:16' },
    { id: 8, title: '漠河舞厅', artist: '柳爽', album: '漠河舞厅', duration: '4:38' },
    { id: 9, title: '错位时空', artist: '艾辰', album: '错位时空', duration: '3:58' },
    { id: 10, title: '踏山河', artist: '是七叔呢', album: '踏山河', duration: '3:22' },
  ];

  const tracks = $derived<TrackItem[]>(
    songs.map((song) => ({
      id: song.id,
      title: song.title,
      artist: song.artist,
      album: song.album,
      duration: song.duration,
    })),
  );

  let currentSongId = $state<number | null>(null);
  let isPlaying = $state(false);

  function playTrack(track: TrackItem) {
    const id = track.id as number;
    if (currentSongId === id) {
      isPlaying = !isPlaying;
      return;
    }
    currentSongId = id;
    isPlaying = true;
  }
</script>

<div class="home-page">
  <div class="section-header">
    <Heading tag="h2" class="text-2xl font-bold text-gray-800">首页</Heading>
    <ButtonGroup>
      {#each tabs as tab}
        <Button
          color={activeTab === tab.id ? 'green' : 'alternative'}
          onclick={() => (activeTab = tab.id)}
        >
          {tab.label}
        </Button>
      {/each}
    </ButtonGroup>
  </div>

  <TrackList
    {tracks}
    activeId={currentSongId}
    {isPlaying}
    onSelect={playTrack}
  />
</div>

<style>
  .home-page {
    padding: 0;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
    flex-wrap: wrap;
    gap: 16px;
  }
</style>
