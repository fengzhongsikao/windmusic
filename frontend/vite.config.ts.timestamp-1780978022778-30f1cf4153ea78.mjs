import "node:module";
import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import { resolve } from "path";
import.meta.url;
var vite_config_default = defineConfig({
	plugins: [tailwindcss(), svelte()],
	resolve: { alias: { "@": resolve("/Users/hd/development/aaa/windmusic/frontend", "src") } }
});
//#endregion
export { vite_config_default as default };

//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoidml0ZS5jb25maWcuanMiLCJuYW1lcyI6W10sInNvdXJjZXMiOlsiL1VzZXJzL2hkL2RldmVsb3BtZW50L2FhYS93aW5kbXVzaWMvZnJvbnRlbmQvdml0ZS5jb25maWcudHMiXSwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHtkZWZpbmVDb25maWd9IGZyb20gJ3ZpdGUnXG5pbXBvcnQge3N2ZWx0ZX0gZnJvbSAnQHN2ZWx0ZWpzL3ZpdGUtcGx1Z2luLXN2ZWx0ZSdcbmltcG9ydCB0YWlsd2luZGNzcyBmcm9tICdAdGFpbHdpbmRjc3Mvdml0ZSdcbmltcG9ydCB7cmVzb2x2ZX0gZnJvbSAncGF0aCdcblxuLy8gaHR0cHM6Ly92aXRlanMuZGV2L2NvbmZpZy9cbmV4cG9ydCBkZWZhdWx0IGRlZmluZUNvbmZpZyh7XG4gIHBsdWdpbnM6IFt0YWlsd2luZGNzcygpLCBzdmVsdGUoKV0sXG4gIHJlc29sdmU6IHtcbiAgICBhbGlhczoge1xuICAgICAgJ0AnOiByZXNvbHZlKF9fZGlybmFtZSwgJ3NyYycpXG4gICAgfVxuICB9XG59KVxuIl0sIm1hcHBpbmdzIjoiOzs7Ozs7QUFNQSxJQUFBLHNCQUFlLGFBQWE7Q0FDMUIsU0FBUyxDQUFDLFlBQVksR0FBRyxPQUFPLENBQUM7Q0FDakMsU0FBUyxFQUNQLE9BQU8sRUFDTCxLQUFLLFFBQUEsZ0RBQW1CLEtBQUssRUFDL0IsRUFDRjtBQUNGLENBQUMifQ==