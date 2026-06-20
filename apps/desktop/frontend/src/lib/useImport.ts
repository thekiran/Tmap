import { useCallback, useState } from 'react';
import { validateScanJson } from './scan-schema';
import { useScanStore } from '../store/useScanStore';
import { wailsBridge, type ScanMode } from './api/AgentBridge';
import { startScan } from './scan-controller';

export interface ImportError {
  path: string;
  message: string;
}

function pickJsonFile(): Promise<string | null> {
  return new Promise((resolve) => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'application/json,.json';
    input.onchange = () => {
      const file = input.files?.[0];
      if (!file) return resolve(null);
      const reader = new FileReader();
      reader.onload = () => resolve(String(reader.result ?? ''));
      reader.onerror = () => resolve(null);
      reader.readAsText(file);
    };
    input.click();
  });
}

export function useImport() {
  const setReport = useScanStore((state) => state.setReport);
  const setScanError = useScanStore((state) => state.setScanError);
  const [errors, setErrors] = useState<ImportError[] | null>(null);
  const [busy, setBusy] = useState(false);

  const loadText = useCallback((text: string, source: 'import' | 'agent' | 'demo' = 'import') => {
    const result = validateScanJson(text);
    if (!result.ok) {
      setErrors(result.errors);
      return false;
    }
    setErrors(null);
    setReport(result.data, source === 'demo' ? 'import' : source);
    return true;
  }, [setReport]);

  const importViaDialog = useCallback(async () => {
    setBusy(true);
    try {
      const text = await wailsBridge.importReport();
      // Fallback to browser picker if wails bridge returned null
      const content = text ?? await pickJsonFile();
      if (content) loadText(content, 'import');
    } catch (error) {
      setErrors([{ path: '(import)', message: (error as Error).message }]);
    } finally {
      setBusy(false);
    }
  }, [loadText]);

  // Delegates to the non-blocking, event-driven scan controller so the existing
  // "Run Scan" buttons keep the live dashboard visible instead of blocking. The
  // controller owns the lifecycle/status; this just kicks it off.
  const runAgent = useCallback(async (mode: ScanMode = 'full', iface = '') => {
    setBusy(true);
    setErrors(null);
    setScanError(null);
    try {
      await startScan(mode, iface);
    } finally {
      setBusy(false);
    }
  }, [setScanError]);

  return {
    busy,
    errors,
    loadText,
    importViaDialog,
    runAgent,
    clearErrors: () => setErrors(null),
  };
}
