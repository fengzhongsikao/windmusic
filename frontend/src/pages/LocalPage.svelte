<script lang="ts">
  import { FolderOpen, Music, Play, Folder } from '@lucide/svelte';

  const localFolders = [
    { id: 1, name: '我的音乐', count: 128 },
    { id: 2, name: '下载', count: 45 },
    { id: 3, name: '录音', count: 12 },
  ];

  const localSongs = [
    { id: 1, title: '晴天', artist: '周杰伦', format: 'MP3', size: '8.2 MB' },
    { id: 2, title: '稻香', artist: '周杰伦', format: 'FLAC', size: '32.5 MB' },
    { id: 3, title: '夜曲', artist: '周杰伦', format: 'MP3', size: '7.8 MB' },
  ];
</script>

<div class="local-page">
  <h2 class="section-title">
    <FolderOpen size={28} />
    本地音乐
  </h2>

  <div class="folder-section">
    <h3 class="sub-title">音乐文件夹</h3>
    <div class="folder-grid">
      {#each localFolders as folder}
        <button class="folder-card">
          <Folder size={32} />
          <div class="folder-info">
            <div class="folder-name">{folder.name}</div>
            <div class="folder-count">{folder.count} 首</div>
          </div>
        </button>
      {/each}
    </div>
  </div>

  <div class="song-section">
    <h3 class="sub-title">本地歌曲</h3>
    <div class="song-list">
      {#each localSongs as song, index}
        <button class="song-row">
          <span class="col-index">{index + 1}</span>
          <span class="col-title">
            <span class="song-cover">
              <Music size={14} />
            </span>
            <span class="song-name">{song.title}</span>
          </span>
          <span class="col-artist">{song.artist}</span>
          <span class="col-format">{song.format}</span>
          <span class="col-size">{song.size}</span>
        </button>
      {/each}
    </div>
  </div>
</div>

<style>
  .local-page {
    padding: 0;
  }

  .section-title {
    font-size: 24px;
    font-weight: 700;
    color: #333;
    margin-bottom: 24px;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .sub-title {
    font-size: 16px;
    font-weight: 600;
    color: #666;
    margin-bottom: 16px;
  }

  .folder-section {
    margin-bottom: 32px;
  }

  .folder-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 16px;
  }

  .folder-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    background: #f5f5f5;
    border: none;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
    color: #667eea;
  }

  .folder-card:hover {
    background: #eee;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  }

  .folder-info {
    text-align: left;
  }

  .folder-name {
    font-size: 14px;
    font-weight: 600;
    color: #333;
  }

  .folder-count {
    font-size: 12px;
    color: #999;
  }

  .song-list {
    border-radius: 12px;
    overflow: hidden;
    background: #fafafa;
  }

  .song-row {
    display: grid;
    grid-template-columns: 50px 1fr 140px 80px 80px;
    padding: 12px 16px;
    border: none;
    background: transparent;
    color: #666;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
    width: 100%;
    align-items: center;
  }

  .song-row:nth-child(even) {
    background: rgba(0, 0, 0, 0.01);
  }

  .song-row:hover {
    background: rgba(0, 0, 0, 0.04);
  }

  .col-index {
    font-size: 13px;
    color: #ccc;
    text-align: center;
  }

  .col-title {
    display: flex;
    align-items: center;
    gap: 12px;
    overflow: hidden;
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

  .song-name {
    font-weight: 500;
    color: #333;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-artist {
    color: #666;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .col-format {
    font-size: 12px;
    color: #667eea;
    background: rgba(102, 126, 234, 0.1);
    padding: 2px 8px;
    border-radius: 4px;
  }

  .col-size {
    color: #999;
    font-size: 13px;
    text-align: right;
  }
</style>
