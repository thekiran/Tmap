/// <reference types="vite/client" />

// Minimal typing for the Wails-bound Go methods exposed on window at runtime.
// In a generated Wails project these come from `frontend/wailsjs/go/main/App`.
// We declare them loosely so the app compiles with or without that codegen.
export {};

declare global {
  interface Window {
    go?: {
      main: {
        App: {
          OpenScanFile(): Promise<string>;
          SaveExport(suggestedName: string, content: string): Promise<string>;
          ListInterfaces(): Promise<string>;
          RunScan(mode: string, iface: string): Promise<string>;
          Platform(): Promise<string>;
        };
      };
    };
  }
}
