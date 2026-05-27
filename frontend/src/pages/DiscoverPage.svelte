<script lang="ts">
  import {
    Button,
    ButtonGroup,
    Heading,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
  } from 'flowbite-svelte';
  import { Music, Play } from '@lucide/svelte';

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

  let currentSong: Song | null = $state(null);
  let isPlaying = $state(false);

  function playSong(song: Song) {
    if (currentSong?.id === song.id) {
      isPlaying = !isPlaying;
    } else {
      currentSong = song;
      isPlaying = true;
    }
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

  <Table hoverable striped class="song-table">
    <TableHead>
      <TableHeadCell class="w-12 text-center">#</TableHeadCell>
      <TableHeadCell>标题</TableHeadCell>
      <TableHeadCell>歌手</TableHeadCell>
      <TableHeadCell>专辑</TableHeadCell>
      <TableHeadCell class="text-right">时长</TableHeadCell>
    </TableHead>
    <TableBody>
      {#each songs as song, index}
        <TableBodyRow
          class="song-row cursor-pointer {currentSong?.id === song.id ? 'playing' : ''}"
          onclick={() => playSong(song)}
        >
          <TableBodyCell class="text-center text-gray-400">
            {#if currentSong?.id === song.id && isPlaying}
              <span class="playing-indicator" aria-hidden="true">
                <span></span><span></span><span></span>
              </span>
            {:else}
              {index + 1}
            {/if}
          </TableBodyCell>
          <TableBodyCell>
            <span class="flex items-center gap-3">
              <span class="song-cover">
                {#if currentSong?.id === song.id && isPlaying}
                  <Play size={14} />
                {:else}
                  <Music size={14} />
                {/if}
              </span>
              <span class="font-medium text-gray-800">{song.title}</span>
            </span>
          </TableBodyCell>
          <TableBodyCell class="text-gray-600">{song.artist}</TableBodyCell>
          <TableBodyCell class="text-gray-500">{song.album}</TableBodyCell>
          <TableBodyCell class="text-right text-gray-400">{song.duration}</TableBodyCell>
        </TableBodyRow>
      {/each}
    </TableBody>
  </Table>
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

  :global(.song-table) {
    border-radius: 12px;
    overflow: hidden;
  }

  :global(.song-row.playing) {
    background: rgba(102, 126, 234, 0.08) !important;
    color: #667eea;
  }

  .song-cover {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f0f0f0;
    border-radius: 6px;
    flex-shrink: 0;
    color: #999;
  }

  :global(.song-row.playing) .song-cover {
    color: #667eea;
    background: rgba(102, 126, 234, 0.1);
  }

  .playing-indicator {
    display: inline-flex;
    align-items: flex-end;
    gap: 2px;
    height: 14px;
  }

  .playing-indicator span {
    width: 3px;
    background: #667eea;
    border-radius: 1px;
    animation: bounce 0.8s ease-in-out infinite;
  }

  .playing-indicator span:nth-child(1) {
    height: 60%;
    animation-delay: 0s;
  }

  .playing-indicator span:nth-child(2) {
    height: 100%;
    animation-delay: 0.2s;
  }

  .playing-indicator span:nth-child(3) {
    height: 40%;
    animation-delay: 0.4s;
  }

  @keyframes bounce {
    0%, 100% { transform: scaleY(1); }
    50% { transform: scaleY(0.4); }
  }
</style>
