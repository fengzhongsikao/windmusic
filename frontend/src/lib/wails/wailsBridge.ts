const WAILS_CALLBACK_NOT_REGISTERED = /Callback 'main\.App\./i;

export function isWailsBridgeReady(): boolean {
  if (typeof window === 'undefined') {
    return false;
  }
  const app = (window as Window & { go?: { main?: { App?: Record<string, unknown> } } }).go?.main
    ?.App;
  return app != null && typeof app.Search === 'function';
}

export function waitForWailsBridge(timeoutMs = 10_000): Promise<boolean> {
  if (isWailsBridgeReady()) {
    return Promise.resolve(true);
  }

  return new Promise((resolve) => {
    const started = performance.now();
    const tick = () => {
      if (isWailsBridgeReady()) {
        resolve(true);
        return;
      }
      if (performance.now() - started >= timeoutMs) {
        resolve(false);
        return;
      }
      requestAnimationFrame(tick);
    };
    tick();
  });
}

export function isWailsCallbackStaleError(err: unknown): boolean {
  const message = err instanceof Error ? err.message : String(err);
  return WAILS_CALLBACK_NOT_REGISTERED.test(message);
}

let reloadScheduled = false;

/** 开发模式下桥接失效时提示并整页重载一次，重新绑定 Go 回调。 */
export function recoverFromStaleWailsBridge(err: unknown): boolean {
  if (!isWailsCallbackStaleError(err)) {
    return false;
  }

  console.error(
    '[Wails] Go 回调未注册，通常是 wails dev 重启或热更新导致前后端不同步。请完全重启 wails dev。',
    err,
  );

  if (import.meta.env.DEV && !reloadScheduled) {
    reloadScheduled = true;
    window.setTimeout(() => {
      window.location.reload();
    }, 300);
  }

  return true;
}
