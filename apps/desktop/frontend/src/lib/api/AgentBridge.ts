export type ScanMode = 'quick' | 'standard' | 'deep' | 'full';

export interface NetworkInterface {
  name: string;
  mac?: string;
  up: boolean;
  loopback: boolean;
  virtual: boolean;
  selected: boolean;
  cidr?: string;
  addresses?: { ip: string; version: number; cidr?: string }[];
}

export interface AgentBridge {
  /**
   * Lists the host's network interfaces so the user can pick one to scan.
   */
  listInterfaces(): Promise<NetworkInterface[]>;

  /**
   * Triggers a new network scan on the Go backend, optionally pinned to a
   * specific interface name.
   */
  runScan(mode?: ScanMode, iface?: string): Promise<string>;

  /**
   * Prompts the user to select a JSON report file to import.
   * Resolves with the raw JSON string if successful.
   */
  importReport(): Promise<string | null>;

  /**
   * Generates a safe-share export from the backend.
   */
  exportReport(): Promise<string>;

  /**
   * Prompts the user to save content to a file.
   */
  saveExport(filename: string, content: string): Promise<void>;

  /**
   * Registers a callback for scan log updates.
   */
  onScanLog(callback: (log: any) => void): () => void;

  /**
   * Registers a callback for scan completion.
   */
  onScanComplete(callback: (reportRaw: string) => void): () => void;
}

// Ensure the global window object knows about the Wails bindings
declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          ListInterfaces?: () => Promise<string>;
          RunScan?: (mode: string, iface: string) => Promise<string>;
          ImportReport?: () => Promise<string>;
          ExportReport?: () => Promise<string>;
          SaveExport?: (filename: string, content: string) => Promise<void>;
        };
      };
    };
    runtime?: {
      EventsOn?: (eventName: string, callback: (...args: any[]) => void) => void;
      EventsOff?: (eventName: string) => void;
    };
  }
}

export const wailsBridge: AgentBridge = {
  listInterfaces: async () => {
    if (window.go?.main?.App?.ListInterfaces) {
      const raw = await window.go.main.App.ListInterfaces();
      try {
        const parsed = JSON.parse(raw);
        return Array.isArray(parsed) ? (parsed as NetworkInterface[]) : [];
      } catch {
        return [];
      }
    }
    console.warn('Wails binding ListInterfaces not found. Running in browser?');
    return [];
  },

  runScan: async (mode: ScanMode = 'standard', iface = '') => {
    if (window.go?.main?.App?.RunScan) {
      return await window.go.main.App.RunScan(mode, iface);
    }
    console.warn('Wails binding RunScan not found. Running in browser?');
    return '';
  },

  importReport: async () => {
    if (window.go?.main?.App?.ImportReport) {
      return await window.go.main.App.ImportReport();
    }
    console.warn('Wails binding ImportReport not found. Running in browser?');
    return null;
  },

  exportReport: async () => {
    if (window.go?.main?.App?.ExportReport) {
      return await window.go.main.App.ExportReport();
    }
    console.warn('Wails binding ExportReport not found. Running in browser?');
    return '{}';
  },

  saveExport: async (filename: string, content: string) => {
    if (window.go?.main?.App?.SaveExport) {
      await window.go.main.App.SaveExport(filename, content);
    } else {
      // Fallback for browser testing
      const blob = new Blob([content], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = filename;
      anchor.click();
      URL.revokeObjectURL(url);
    }
  },

  onScanLog: (callback) => {
    if (window.runtime?.EventsOn) {
      window.runtime.EventsOn('scan-log', callback);
      return () => {
        if (window.runtime?.EventsOff) {
          window.runtime.EventsOff('scan-log');
        }
      };
    }
    return () => {};
  },

  onScanComplete: (callback) => {
    if (window.runtime?.EventsOn) {
      window.runtime.EventsOn('scan-complete', callback);
      return () => {
         if (window.runtime?.EventsOff) {
          window.runtime.EventsOff('scan-complete');
        }
      };
    }
    return () => {};
  }
};
