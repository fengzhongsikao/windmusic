# README

## 项目说明

这是官方的 Wails + Svelte-TS 模板项目。

## 开发模式

在项目目录执行 `wails dev` 即可进入实时开发模式。该命令会启动 Vite 开发服务器，
前端改动可以快速热更新。

如果你希望在浏览器中开发并调用 Go 方法，还可以访问 http://localhost:34115。
在浏览器打开该地址后，可通过开发者工具调用后端 Go 代码。

## 构建

执行 `wails build` 可构建可分发的生产版本应用。

## JS 源文件保存位置

导入的音源脚本（`*.js`）会被复制到当前用户的配置目录中。

- 根目录：`os.UserConfigDir()/windmusic`
- 脚本文件：`os.UserConfigDir()/windmusic/sources/*.js`
- 源索引元数据：`os.UserConfigDir()/windmusic/sources.json`

各平台常见路径如下：

- macOS
  - `~/Library/Application Support/windmusic/sources/`
  - `~/Library/Application Support/windmusic/sources.json`
- Windows
  - `%AppData%\windmusic\sources\`
  - `%AppData%\windmusic\sources.json`
- Linux
  - `$XDG_CONFIG_HOME/windmusic/sources/`（当设置了 `XDG_CONFIG_HOME` 时）
  - `~/.config/windmusic/sources/`（默认）
  - `sources.json` 位于同级 `windmusic` 目录中
