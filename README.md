# WindMusic

基于 [Wails v2](https://wails.io/) 的桌面音乐客户端：前端负责界面与音频播放，Go 后端负责 Meting 网络源的数据拉取，以及收藏、最近播放等本地持久化。

## 功能概览

| 模块 | 说明 |
|------|------|
| **首页发现** | 按分类 Tab 拉取推荐曲目（走 `Search` API，带缓存） |
| **搜索** | 关键词搜索、分页、结果缓存；支持从顶栏全局搜索进入 |
| **播放** | 底部播放条：播放/暂停、进度、音量、上一首/下一首、随机/循环、播放队列 |
| **歌词** | 沉浸式视图与 LRC 同步（`GetLyric`） |
| **我喜欢的音乐** | 本地 `favorites.json`，支持编辑、全选、批量删除 |
| **最近播放** | 本地 `recent.json`，开始播放时自动记录 |
| **设置** | 配置多个 Meting API 节点，可添加、删除、切换当前节点 |
| **占位页面** | 推荐歌单、排行榜、本地音乐等为 UI 占位，尚未接入后端 |

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面壳 | Wails v2、Go 1.23 |
| 前端 | Svelte 5、TypeScript、Vite 8、Tailwind CSS 4 |
| 路由 | svelte-spa-router（Hash 路由） |
| UI | Lucide 图标、Skeleton UI 部分组件 |
| 包管理 | Bun（`wails.json` 中配置） |

## 架构

```
┌─────────────────────────────────────────────────────────┐
│  Svelte 前端                                             │
│  · 页面 / 组件 / 全局 stores                             │
│  · HTMLAudioElement 实际解码与播放（audioEngine）          │
│  · 内存 $state；数据由 Go 经 Events 推送（无浏览器持久化）    │
└───────────────────────┬─────────────────────────────────┘
                        │ Wails Bind（wailsjs/go/main/App）
┌───────────────────────▼─────────────────────────────────┐
│  Go 后端（app.go → music/ → internal/musicsearch/）       │
│  · Search / GetMusicURL / GetLyric / GetPic               │
│  · 设置、Meting、收藏、歌单、本地库等（JSON + 事件推送）     │
└───────────────────────┬─────────────────────────────────┘
                        │ HTTP
┌───────────────────────▼─────────────────────────────────┐
│  Meting API 节点（用户可配置多个，如 meting.mikus.ink）    │
└─────────────────────────────────────────────────────────┘
```

**职责边界（重要）**

1. 用户在 UI 选择曲目，前端组装 `sourceId`、`platform`、`metaJson`。
2. 播放地址、歌词、封面由 Go 调用 Meting 相关逻辑后返回。
3. **Go 不解码、不播放音频**；`stores/audioEngine.ts` 中的 `<audio>` 是唯一播放执行者。
4. 队列、随机、循环等状态集中在 `stores/player.svelte.ts`，页面不重复维护播放状态。

**播放数据流**

```
选曲 → player.svelte（队列/当前曲）
     → GetMusicURL → 设置 audio.src → play
     → GetLyric / GetPic（按需）
     → RecordRecent / 收藏状态同步
```

**音源标识**

- Meting 节点：`meting::<baseURL>`（例如 `meting::https://meting.mikus.ink`）
- 未配置 Meting 时，搜索页可能回退 `builtin::network`；收藏封面等能力则要求先在设置中配置节点。

当前搜索实现面向 **网易云**（`platform=netease`）；部分 Meting 公共节点仅支持网易云 `type=search`。

## 目录结构

```
windmusic/
├── main.go                 # Wails 入口，嵌入 frontend/dist
├── app.go                  # 导出给前端的 Go 方法
├── music/                  # 业务层：搜索、收藏、最近播放、路径
│   ├── search.go
│   ├── favorites.go
│   ├── recent.go
│   └── source.go
├── internal/
│   ├── music/              # 共享模型（SongItem、FavoriteSong 等）
│   └── musicsearch/        # Meting HTTP 客户端
├── frontend/
│   ├── src/
│   │   ├── pages/          # 页面（discover、search、library、settings…）
│   │   ├── components/     # TrackList、PlayerBar、SearchBar…
│   │   ├── stores/         # player、audioEngine、lyrics、toast
│   │   └── lib/            # meting、wailsPlayer、playerTrack、lrc…
│   └── wailsjs/            # Wails 生成绑定（勿手改）
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
| `/local` | 本地音乐（占位） |
| `/recommend`、`/ranking` | 推荐/排行榜（占位） |
| `/settings` | Meting 源设置 |

顶栏搜索框右侧齿轮进入设置。

## Go API（Wails 绑定）

| 方法 | 作用 |
|------|------|
| `Search` | 按音源、平台、关键词分页搜索 |
| `GetMusicURL` | 从 `metaJson` 解析播放地址 |
| `GetLyric` | 拉取 LRC 歌词 |
| `GetPic` | 获取封面 URL |
| `GetSourceDataDir` | 返回应用数据根目录 |
| `ListFavorites` / `AddFavorite` / `RemoveFavorite` / `IsFavorite` | 收藏 |
| `ListRecent` / `RecordRecent` / `RemoveRecent` / `ClearRecent` | 最近播放 |

## 数据存储

### 应用数据目录（Go 读写）

根目录：`os.UserConfigDir()/windmusic`（应用内 **设置** 页也会显示 `GetSourceDataDir()` 返回的完整路径）

| 文件 | 内容 |
|------|------|
| `favorites.json` | 收藏列表 |
| `recent.json` | 最近播放 |
| `playlists.json` | 自建歌单 |
| `local-folders.json` | 本地音乐扫描的文件夹列表 |
| `local-scan-cache.json` | 本地扫描结果缓存 |
| `local-scan-extras.json` | 本地封面、歌词等扩展数据 |

各平台常见路径：

- **macOS**：`~/Library/Application Support/windmusic/`
- **Windows**：`%AppData%\windmusic\`
- **Linux**：`~/.config/windmusic/`（或 `$XDG_CONFIG_HOME/windmusic/`）

查看该目录下的 JSON 文件：

**macOS**

```bash
ls -la ~/Library/Application\ Support/windmusic/
```

**Linux**

```bash
ls -la ~/.config/windmusic/
```

若设置了 `XDG_CONFIG_HOME`：

```bash
ls -la "${XDG_CONFIG_HOME}/windmusic/"
```

**Windows（PowerShell）**

```powershell
Get-ChildItem "$env:APPDATA\windmusic"
```

**Windows（命令提示符 cmd）**

```cmd
dir "%APPDATA%\windmusic"
```

应用启动时会清除旧版 `localStorage` 键（`windmusic:meting-*`、`windmusic:player-settings`），不再读取其中数据。

## 环境要求

- [Go](https://go.dev/) 1.23+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- [Bun](https://bun.sh/)（安装前端依赖与构建）

## 开发

```bash
# 安装前端依赖
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

产物为当前平台可执行文件；前端会先执行 `bun run build` 生成 `frontend/dist`，再由 Go `embed` 打包进二进制。

## 相关文档

- [AGENTS.md](./AGENTS.md) — 仓库约定、校验清单、架构规则（供开发与 AI 助手参考）
- [Wails 文档](https://wails.io/docs/introduction)
