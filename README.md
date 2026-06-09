# WindMusic

基于 [Wails v2](https://wails.io/) 的跨平台桌面音乐客户端。前端负责界面与音频播放，Go 后端负责 Meting 网络源数据拉取、本地音乐库扫描（SQLite 持久化），以及收藏、歌单、最近播放等 JSON 持久化。

前端使用 **Bun** 作为 JavaScript 运行时与包管理器（见 `wails.json` 中的 `frontend:install` / `frontend:build` 配置）。

## 演示

[B 站演示视频](https://www.bilibili.com/video/BV1L8V36bEFW/)

## 截图

| 首页发现 | 本地音乐 | 最近播放 |
|:---:|:---:|:---:|
| ![首页发现](./images/home.png) | ![本地音乐](./images/local.png) | ![最近播放](./images/recent.png) |

## 功能概览

| 模块 | 说明 |
|------|------|
| **首页发现** | 推荐 / 华语 / 流行 / 摇滚 / 电子等分类 Tab，通过 Meting 搜索拉取曲目，Go 端内存缓存 |
| **搜索** | 关键词搜索、分页；顶栏全局搜索框快捷进入 |
| **播放** | 底部播放条：播放/暂停、进度、音量、上一首/下一首、随机/循环、播放队列 |
| **沉浸式歌词** | 歌曲详情页：频谱可视化 + LRC 歌词同步滚动，点击歌词可跳转 |
| **我喜欢的音乐** | 收藏管理，支持编辑、全选、批量删除 |
| **最近播放** | 开始播放时自动记录，支持清空与编辑 |
| **自建歌单** | 创建、删除歌单，向歌单添加/移除歌曲，侧边栏快速访问 |
| **本地音乐** | 添加/移除文件夹、扫描 MP3/FLAC/M4A/AAC/OGG/WAV/WMA 等格式、文件夹别名、按目录 Tab 筛选、读取内嵌封面与歌词；扫描结果与元数据存于 SQLite |
| **设置** | 配置多个 Meting API 节点，添加、删除、切换当前节点；显示应用数据目录 |
| **占位页面** | 推荐歌单、排行榜页面已预留路由，尚未接入后端 |

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面壳 | Wails v2、Go 1.25 |
| 本地库持久化 | [SQLite](https://www.sqlite.org/)（[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)，纯 Go、无 CGO） |
| 音频元数据 | [dhowden/tag](https://github.com/dhowden/tag)、[mewkiz/flac](https://github.com/mewkiz/flac)、[tcolgate/mp3](https://github.com/tcolgate/mp3) |
| 前端 | Svelte 5、TypeScript、Vite 8、Tailwind CSS 4 |
| 路由 | svelte-spa-router（Hash 路由） |
| UI | Lucide 图标、Skeleton UI、[@tanstack/svelte-virtual](https://tanstack.com/virtual)（大列表虚拟滚动） |
| JS 运行时 / 包管理 | [Bun](https://bun.sh/) |

## 架构

```
┌─────────────────────────────────────────────────────────┐
│  Svelte 前端                                             │
│  · 页面 / 组件 / 全局 stores                             │
│  · HTMLAudioElement 实际解码与播放（audioEngine）          │
│  · 数据由 Go 经 Wails Events 推送，持久化在后端            │
└───────────────────────┬─────────────────────────────────┘
                        │ Wails Bind（wailsjs/go/main/App）
┌───────────────────────▼─────────────────────────────────┐
│  Go 后端（app*.go → music/ → internal/）                  │
│  · Search / GetMusicURL / GetLyric / GetPic               │
│  · 本地库扫描（SQLite）、Meting 源、收藏/歌单（JSON）等    │
└───────────────────────┬─────────────────────────────────┘
                        │ HTTP / 本地文件 / SQLite
┌───────────────────────▼─────────────────────────────────┐
│  Meting API 节点 · 本地音频文件夹 · local-library.db       │
└─────────────────────────────────────────────────────────┘
```

**职责边界**

1. 用户在 UI 选择曲目，前端组装 `sourceId`、`platform`、`metaJson`。
2. 在线播放地址、歌词、封面由 Go 调用 Meting 逻辑后返回；本地曲目通过 `GetLocalAudioStream` 提供可播放 URL。
3. **Go 不解码、不播放音频**；`stores/playback/audioEngine.ts` 中的 `<audio>` 是唯一播放执行者。
4. 队列、随机、循环等状态集中在 `stores/playback/player.svelte.ts`。

**播放数据流**

```
选曲 → player.svelte（队列/当前曲）
     → GetMusicURL / GetLocalAudioStream → 设置 audio.src → play
     → GetLyric / GetPic / GetLocalSongExtras（按需）
     → RecordRecent / 收藏状态同步
```

**音源标识**

- Meting 节点：`meting::<baseURL>`（例如 `meting::https://meting.mikus.ink`）
- 本地文件：`local` 平台，通过文件路径标识
- 未配置 Meting 时，搜索页可能回退 `builtin::network`；收藏封面等能力则要求先在设置中配置节点

当前 Meting 搜索面向 **网易云**（`platform=netease`）。

## 目录结构

```
windmusic/
├── main.go                 # Wails 入口，嵌入 frontend/dist
├── app.go                  # 核心导出方法（搜索、URL、歌词、封面）
├── app_library.go          # 收藏、最近播放、歌单
├── app_local.go            # 本地音乐库
├── app_settings.go         # Meting / 播放器 / 发现页缓存
├── app_sync.go             # 启动时数据推送与事件 emit
├── music/                  # 后端业务层
│   ├── appdata/            # 应用数据目录
│   ├── cache/              # 内存缓存（发现页推荐）
│   ├── events/             # Wails 事件名常量
│   ├── local/              # 本地文件夹扫描、SQLite 缓存、封面文件
│   ├── meting/             # Meting 搜索与 URL 解析
│   └── persist/            # JSON 持久化（收藏、歌单、设置等）
├── internal/
│   ├── music/              # 共享模型（SongItem、LocalSong 等）
│   └── musicsearch/        # Meting HTTP 客户端
├── frontend/               # Svelte 5 前端（Bun 管理依赖）
│   ├── src/
│   │   ├── pages/          # 页面（discover、search、library、settings…）
│   │   ├── components/     # 可复用 UI 组件
│   │   ├── stores/         # 全局状态（playback、library、sources、sync）
│   │   └── lib/            # 工具与 Wails 封装
│   └── wailsjs/            # Wails 生成绑定（勿手改）
├── images/                 # 软件截图（README 展示用）
├── build/                  # 各平台打包资源
├── wails.json
└── AGENTS.md               # 面向 AI 助手的开发约定
```

## 前端路由

| 路径 | 页面 |
|------|------|
| `/`、`/discover` | 首页发现 |
| `/search` | 搜索（`#search?q=关键词&page=1`） |
| `/favorites` | 我喜欢的音乐 |
| `/recent` | 最近播放 |
| `/local` | 本地音乐 |
| `/playlist/:id` | 自建歌单详情 |
| `/recommend`、`/ranking` | 推荐/排行榜（占位） |
| `/settings` | Meting 源与数据目录 |

顶栏搜索框右侧齿轮进入设置。

## 环境要求

- [Go](https://go.dev/) 1.25+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- [Bun](https://bun.sh/) — 前端依赖安装、开发服务器与构建（`wails.json` 已配置）

## 开发

```bash
# 安装前端依赖（Bun）
bun install --cwd frontend

# 推荐：完整桌面应用热重载
wails dev
```

仅前端（不启动 Wails 窗口时）：

```bash
bun run --cwd frontend dev
bun run --cwd frontend check   # svelte-check 类型检查
```

`wails dev` 启动后，也可在浏览器访问 Vite 开发地址（终端会打印，常见为 `http://localhost:34115`）调试前端；调用 Go 方法仍需 Wails 运行时。

**首次使用前**：进入 **设置**，添加至少一个 Meting API 地址（如 `https://meting.mikus.ink`），并点击设为当前。

## 构建

```bash
wails build
```

产物为当前平台可执行文件。构建时 Wails 会先在前端目录执行 `bun install` 与 `bun run build` 生成 `frontend/dist`，再由 Go `embed` 打包进二进制。

## 发布（GitHub Releases）

修改 `wails.json` 中的 `info.productVersion` 并推送到 `main` 分支后，GitHub Actions 会自动检测版本变更、在 macOS / Windows / Linux 上构建，并创建对应 Release。

```bash
# 1. 编辑 wails.json → info.productVersion（例如 1.0.5）
# 2. 提交并推送
git add wails.json
git commit -m "chore: release v1.0.5"
git push origin main
```

仅当 `productVersion` **相对上一提交发生变化** 时才会打包；若只改了 `wails.json` 的其他字段（如 author），不会触发发布。

| 平台 | 产物 |
|------|------|
| macOS (Universal) | `windmusic-v{version}-macos-universal.zip` |
| Windows (amd64) | `windmusic-amd64-installer.exe`（NSIS 安装包） |
| Linux (amd64) | `windmusic-v{version}-linux-amd64.tar.gz` |

Release 标签为 `v{productVersion}`（如 `v1.0.5`）。也可在 GitHub **Actions → Release → Run workflow** 手动触发，版本号读取当前 `wails.json`。

## 数据存储

所有持久化数据由 Go 后端读写，前端不直接访问磁盘。根目录为 `os.UserConfigDir()/windmusic`（设置页通过 `GetSourceDataDir()` 显示完整路径）。

各平台常见路径：

- **macOS**：`~/Library/Application Support/windmusic/`
- **Windows**：`%AppData%\windmusic\`
- **Linux**：`~/.config/windmusic/`（或 `$XDG_CONFIG_HOME/windmusic/`）

### 文件一览

| 文件 / 目录 | 格式 | 管理模块 | 说明 |
|-------------|------|----------|------|
| `favorites.json` | JSON | `music/persist/favorites.go` | 收藏列表 |
| `recent.json` | JSON | `music/persist/recent.go` | 最近播放 |
| `playlists.json` | JSON | `music/persist/playlists.go` | 自建歌单 |
| `player-settings.json` | JSON | `music/persist/player_settings.go` | 播放器 UI 偏好 |
| `meting-settings.json` | JSON | `music/persist/meting_settings.go` | Meting API 节点配置 |
| `local-folders.json` | JSON | `music/local/library.go` | 本地音乐文件夹列表与别名 |
| `local-library.db` | SQLite | `music/local/db.go` | 本地扫描缓存与歌曲扩展元数据 |
| `local-covers/` | 图片文件 | `music/local/coverfile.go` | 本地曲目内嵌封面（按哈希键去重存储） |

发现页推荐结果仅保存在 Go 进程内存中（`music/cache/discover.go`，TTL 5 分钟），**不会**写入磁盘。

### JSON 文件字段说明

#### `favorites.json`

对象数组，每项为一条收藏记录（`FavoriteSong`）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 曲目 ID（在线为平台 song id，本地为文件路径） |
| `title` | string | 标题 |
| `artist` | string | 艺术家 |
| `album` | string | 专辑（可选） |
| `duration` | string | 时长显示文本（可选） |
| `coverUrl` | string | 封面 URL 或 data URL（可选） |
| `sourceId` | string | 音源标识，如 `meting::https://…`（可选） |
| `platform` | string | 平台，如 `netease`、`local`（可选） |
| `metaJson` | string | 在线播放所需的原始元数据 JSON 字符串（可选） |

#### `recent.json`

对象数组，每项为一条最近播放记录（`RecentSong`），字段与收藏基本相同，另增：

| 字段 | 类型 | 说明 |
|------|------|------|
| `playedAt` | string (RFC3339) | 最近一次开始播放的时间 |

#### `playlists.json`

对象数组，每项为一个自建歌单（`UserPlaylist`）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 歌单唯一 ID |
| `name` | string | 歌单名称 |
| `createdAt` | string (RFC3339) | 创建时间 |
| `songs` | array | 歌单内曲目列表，元素结构与 `favorites.json` 相同 |

#### `player-settings.json`

单个对象（`PlayerSettings`），保存播放器与详情页 UI 偏好：

| 字段 | 类型 | 说明 |
|------|------|------|
| `volume` | number | 音量 0–100 |
| `muted` | boolean | 是否静音 |
| `repeatMode` | string | 循环模式：`off` / `all` / `one` |
| `shuffled` | boolean | 是否开启随机播放 |
| `waveformSpread` | string | 频谱展开方向：`center-out` / `right-left` |
| `detailHideLyrics` | boolean | 详情页是否隐藏歌词 |
| `detailHideVisual` | boolean | 详情页是否隐藏频谱 |
| `detailCoverShape` | string | 封面形状：`round` / `square` |
| `detailCoverSpin` | boolean | 播放时封面是否旋转 |
| `detailHidePlayerBar` | boolean | 详情页是否隐藏底部播放条 |

#### `meting-settings.json`

单个对象（`MetingSettings`），Meting API 节点配置：

| 字段 | 类型 | 说明 |
|------|------|------|
| `urls` | string[] | 已添加的 Meting API 根地址列表 |
| `activeUrl` | string | 当前使用的节点地址 |
| `platform` | string | 搜索平台，目前固定为 `netease`（网易云） |

#### `local-folders.json`

单个对象，记录用户添加的本地音乐目录：

| 字段 | 类型 | 说明 |
|------|------|------|
| `paths` | string[] | 已添加的文件夹绝对路径列表 |
| `aliases` | object | 可选，键为文件夹路径、值为用户自定义显示名称 |

扫描结果本身**不**存在此文件中，而是写入 `local-library.db`。

### SQLite：`local-library.db`

由 `music/local/db.go` 管理，启用 WAL 模式（`journal_mode=WAL`），单连接写入。

#### 表 `scan_entries`

本地音频文件扫描缓存，按文件路径索引。文件未修改时可跳过重复解析。

| 列 | 类型 | 说明 |
|----|------|------|
| `path` | TEXT (PK) | 音频文件绝对路径 |
| `mod_time_unix` | INTEGER | 文件最后修改时间（Unix 秒），用于判断是否需要重新扫描 |
| `song_json` | TEXT | `LocalSong` 对象的 JSON 字符串 |

`song_json` 内字段（`LocalSong`）：

| 字段 | 说明 |
|------|------|
| `id` | 与 `filePath` 相同 |
| `title` | 从 ID3/Vorbis 等标签读取，缺省为文件名 |
| `artist` | 艺术家，缺省为「未知艺术家」 |
| `album` | 专辑（可选） |
| `duration` | 格式化时长字符串 |
| `filePath` | 文件绝对路径 |
| `format` | 扩展名（如 `.mp3`、`.flac`） |
| `size` | 人类可读的文件大小 |

列表接口返回的 `LocalSong` **不含**封面与歌词大字段；这两项按需从 `song_extras` 加载。

#### 表 `song_extras`

按文件路径存储体积较大的扩展元数据，与 `scan_entries` 一一对应（同一路径）。

| 列 | 类型 | 说明 |
|----|------|------|
| `path` | TEXT (PK) | 音频文件绝对路径 |
| `cover_key` | TEXT | 封面哈希键（SHA-256 前 12 字节 hex），指向 `local-covers/` 中的图片文件；多首歌曲共用相同封面时会复用同一 key |
| `lyric` | TEXT | 从音频文件内嵌标签读取的歌词文本（LRC 或纯文本），缺省为空字符串 |

### 目录：`local-covers/`

从内嵌专辑封面提取的二进制图片，以 `{cover_key}{扩展名}` 命名（扩展名由 MIME 推断，常见 `.jpg` / `.png`）。`cover_key` 由封面 data URL 的 SHA-256 哈希生成，相同封面只存一份，节省空间。

扫描时不再使用的旧路径会从 `scan_entries` / `song_extras` 中清理；未被任何曲目引用的封面文件也会被 prune。

## Go API（Wails 绑定）

<details>
<summary>展开完整 API 列表</summary>

| 方法 | 作用 |
|------|------|
| `Search` | 按音源、平台、关键词分页搜索 |
| `GetMusicURL` | 从 `metaJson` 解析在线播放地址 |
| `GetLyric` | 拉取 LRC 歌词 |
| `GetPic` | 获取封面 URL |
| `GetSourceDataDir` | 返回应用数据根目录 |
| `ListFavorites` / `AddFavorite` / `RemoveFavorite` / `IsFavorite` | 收藏 |
| `ListRecent` / `RecordRecent` / `RemoveRecent` / `ClearRecent` | 最近播放 |
| `ListPlaylists` / `CreatePlaylist` / `GetPlaylist` / `DeletePlaylist` | 歌单 |
| `AddPlaylistSong` / `RemovePlaylistSong` | 歌单曲目管理 |
| `PickLocalMusicFolder` / `RemoveLocalMusicFolder` / `SetLocalFolderAlias` | 本地文件夹管理 |
| `ScanLocalLibrary` / `ListLocalLibrary` / `GetLocalFolderSongs` | 本地库扫描与查询 |
| `GetLocalAudioStream` / `GetLocalSongExtras` / `GetLocalSongCovers` | 本地播放与元数据 |
| `GetMetingSettings` / `AddMetingURL` / `RemoveMetingURL` / `SetActiveMetingURL` | Meting 源配置 |
| `GetPlayerSettings` / `UpdatePlayerSettings` | 播放器设置 |
| `GetDiscoverRecommendCache` / `SetDiscoverRecommendCache` | 发现页推荐缓存 |

</details>

## 相关文档

- [AGENTS.md](./AGENTS.md) — 仓库约定、校验清单、架构规则
- [Wails 文档](https://wails.io/docs/introduction)
- [Bun 文档](https://bun.sh/docs)
