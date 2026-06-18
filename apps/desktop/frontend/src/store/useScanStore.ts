import { create } from 'zustand';
import { normalizeScan } from '../lib/normalize-scan';
import { buildTopologyViewModel } from '../lib/build-topology-view-model';
import type { RawScanReport } from '../lib/scan-schema';
import type { NormalizedScanReport } from '../lib/models';

interface ScanState {
  report: RawScanReport | null;
  normalized: NormalizedScanReport | null;
  isScanning: boolean;
  scanProgress: number;
  scanError: string | null;
  logs: string[];

  setReport: (report: RawScanReport, source?: string) => void;
  setScanning: (isScanning: boolean) => void;
  setScanError: (error: string | null) => void;
  addLog: (log: string) => void;
  clearLogs: () => void;
  clearReport: () => void;
}

export const useScanStore = create<ScanState>((set) => ({
  report: null,
  normalized: null,
  isScanning: false,
  scanProgress: 0,
  scanError: null,
  logs: [],

  setReport: (report: RawScanReport, _source?: string) => {
    try {
      const normalizedBase = normalizeScan(report);
      // Ensure topology view model is built and injected
      const topology = buildTopologyViewModel(normalizedBase);
      const normalized: NormalizedScanReport = { ...normalizedBase, topology };
      set({ report, normalized, scanError: null });
    } catch (e) {
      console.error("Failed to normalize report", e);
      set({ report, normalized: null });
    }
  },

  clearReport: () => set({ report: null, normalized: null }),
  setScanning: (isScanning: boolean) => set({ isScanning }),
  setScanError: (scanError: string | null) => set({ scanError }),
  addLog: (log: string) => set((state) => ({ logs: [...state.logs, log] })),
  clearLogs: () => set({ logs: [] }),
}));
