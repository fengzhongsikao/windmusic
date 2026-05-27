package lxruntime

import (
	"fmt"
	"sync"
	"time"

	"windmusic/internal/music"

	"github.com/dop251/goja"
)

const (
	EventInited      = "inited"
	EventRequest     = "request"
	EventUpdateAlert = "updateAlert"
)

type PlatformMeta struct {
	Key       string
	Name      string
	Type      string
	Actions   []string
	Qualities []string
}

type InitResult struct {
	Platforms []PlatformMeta
}

type Runtime struct {
	vm             *goja.Runtime
	mu             sync.Mutex
	rawScript      string
	meta           ScriptMeta
	platforms      []PlatformMeta
	requestHandler goja.Callable
	initErr        error
	initOnce       sync.Once
	initDone       chan struct{}
	http           *httpClient
}

func NewRuntime(rawScript string) *Runtime {
	meta := ParseScriptMeta(rawScript)
	rt := &Runtime{
		rawScript: rawScript,
		meta:      meta,
		initDone:  make(chan struct{}),
	}
	rt.vm = goja.New()
	rt.http = newHTTPClient(rt.vm, &rt.mu)
	rt.setupGlobalLX()
	return rt
}

func (r *Runtime) Init(timeout time.Duration) (*InitResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.vm.RunString(r.rawScript)
	if err != nil {
		r.finishInit(err)
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	select {
	case <-r.initDone:
		if r.initErr != nil {
			return nil, r.initErr
		}
		return &InitResult{Platforms: r.platforms}, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("source initialization timeout")
	}
}

func (r *Runtime) Meta() ScriptMeta {
	return r.meta
}

func (r *Runtime) Platforms() []PlatformMeta {
	return append([]PlatformMeta(nil), r.platforms...)
}

func (r *Runtime) GetMusicURL(source string, musicInfo map[string]interface{}, quality string) (string, error) {
	return r.callRequest(source, "musicUrl", map[string]interface{}{
		"type":      quality,
		"musicInfo": musicInfo,
	})
}

func (r *Runtime) GetLyric(source string, musicInfo map[string]interface{}) (*music.LyricInfo, error) {
	result, err := r.callRequestValue(source, "lyric", map[string]interface{}{
		"musicInfo": musicInfo,
	})
	if err != nil {
		return nil, err
	}

	exported := result.Export()
	obj, ok := exported.(map[string]interface{})
	if !ok {
		if str, ok := exported.(string); ok {
			return &music.LyricInfo{Lyric: str}, nil
		}
		return nil, fmt.Errorf("unexpected lyric response")
	}

	return &music.LyricInfo{
		Lyric:   stringValue(obj["lyric"], ""),
		TLyric:  stringValue(obj["tlyric"], ""),
		RLyric:  stringValue(obj["rlyric"], ""),
		LXLyric: stringValue(obj["lxlyric"], ""),
	}, nil
}

func (r *Runtime) GetPic(source string, musicInfo map[string]interface{}) (string, error) {
	return r.callRequest(source, "pic", map[string]interface{}{
		"musicInfo": musicInfo,
	})
}

func (r *Runtime) callRequest(source, action string, info map[string]interface{}) (string, error) {
	value, err := r.callRequestValue(source, action, info)
	if err != nil {
		return "", err
	}
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return "", fmt.Errorf("empty response for action %s", action)
	}
	return fmt.Sprint(value.Export()), nil
}

func (r *Runtime) callRequestValue(source, action string, info map[string]interface{}) (goja.Value, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.requestHandler == nil {
		return nil, fmt.Errorf("source request handler not registered")
	}

	payload := map[string]interface{}{
		"source": source,
		"action": action,
		"info":   info,
	}

	result, err := r.requestHandler(goja.Undefined(), r.vm.ToValue(payload))
	if err != nil {
		return nil, err
	}

	return awaitPromise(r.vm, result, 60*time.Second)
}

func (r *Runtime) setupGlobalLX() {
	lx := r.vm.NewObject()
	eventNames := r.vm.NewObject()
	_ = eventNames.Set("inited", EventInited)
	_ = eventNames.Set("request", EventRequest)
	_ = eventNames.Set("updateAlert", EventUpdateAlert)
	_ = lx.Set("EVENT_NAMES", eventNames)
	_ = lx.Set("version", "2.0.0")
	_ = lx.Set("env", "desktop")

	scriptInfo := r.vm.NewObject()
	_ = scriptInfo.Set("name", r.meta.Name)
	_ = scriptInfo.Set("description", r.meta.Description)
	_ = scriptInfo.Set("version", r.meta.Version)
	_ = scriptInfo.Set("author", r.meta.Author)
	_ = scriptInfo.Set("homepage", r.meta.Homepage)
	_ = scriptInfo.Set("rawScript", r.rawScript)
	_ = lx.Set("currentScriptInfo", scriptInfo)

	_ = lx.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventName := call.Arguments[0].String()
		handler, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			return goja.Undefined()
		}
		if eventName == EventRequest {
			r.requestHandler = handler
		}
		return goja.Undefined()
	})

	_ = lx.Set("send", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventName := call.Arguments[0].String()
		data := exportObject(call.Arguments[1])
		if eventName == EventInited {
			r.handleInited(data)
		}
		return goja.Undefined()
	})

	_ = lx.Set("request", r.http.request)
	_ = lx.Set("utils", buildUtils(r.vm))

	global := r.vm.GlobalObject()
	_ = global.Set("globalThis", global)
	_ = global.Set("lx", lx)
	_ = global.Set("console", r.buildConsole())

	r.vm.Set("setTimeout", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		fn, ok := goja.AssertFunction(call.Arguments[0])
		if !ok {
			return goja.Undefined()
		}
		delay := int64(0)
		if len(call.Arguments) > 1 {
			delay = call.Arguments[1].ToInteger()
		}
		go func() {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			r.mu.Lock()
			defer r.mu.Unlock()
			_, _ = fn(goja.Undefined())
		}()
		return goja.Undefined()
	})
	_ = r.vm.Set("clearTimeout", func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})
}

func (r *Runtime) handleInited(data map[string]interface{}) {
	if status, ok := data["status"].(bool); ok && !status {
		r.finishInit(fmt.Errorf("source initialization rejected"))
		return
	}

	sources := mapValue(data["sources"])
	platforms := make([]PlatformMeta, 0, len(sources))
	for key, raw := range sources {
		item := mapValue(raw)
		actions := stringSlice(item["actions"])
		qualities := stringSlice(item["qualitys"])
		if len(qualities) == 0 {
			qualities = stringSlice(item["qualities"])
		}
		platforms = append(platforms, PlatformMeta{
			Key:       key,
			Name:      stringValue(item["name"], key),
			Type:      stringValue(item["type"], "music"),
			Actions:   actions,
			Qualities: qualities,
		})
	}
	r.platforms = platforms
	r.finishInit(nil)
}

func (r *Runtime) finishInit(err error) {
	r.initOnce.Do(func() {
		r.initErr = err
		close(r.initDone)
	})
}

func (r *Runtime) buildConsole() *goja.Object {
	consoleObj := r.vm.NewObject()
	logFn := func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	}
	_ = consoleObj.Set("log", logFn)
	_ = consoleObj.Set("info", logFn)
	_ = consoleObj.Set("warn", logFn)
	_ = consoleObj.Set("error", logFn)
	_ = consoleObj.Set("group", logFn)
	_ = consoleObj.Set("groupEnd", logFn)
	return consoleObj
}

func stringSlice(value interface{}) []string {
	raw, ok := value.([]interface{})
	if !ok {
		if arr, ok := value.([]string); ok {
			return arr
		}
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		result = append(result, fmt.Sprint(item))
	}
	return result
}

func ToMusicInfoMap(info music.MusicInfo) map[string]interface{} {
	if info.Raw != nil {
		return info.Raw
	}
	result := map[string]interface{}{}
	if info.SongMID != "" {
		result["songmid"] = info.SongMID
	}
	if info.Hash != "" {
		result["hash"] = info.Hash
	}
	if info.Name != "" {
		result["name"] = info.Name
	}
	if info.Singer != "" {
		result["singer"] = info.Singer
	}
	if info.AlbumName != "" {
		result["albumName"] = info.AlbumName
	}
	return result
}
