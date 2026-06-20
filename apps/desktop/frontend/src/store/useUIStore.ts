import { create } from 'zustand';
import type { LayoutEngine } from '../lib/models';
import type { PacketIntensity } from '../lib/packet-flow';

export type ScreenId = 'overview' | 'topology' | 'devices' | 'evidence' | 'probes' | 'export';
const layoutStorageKey = 'iad.topology.layoutPositions.v2';
const layoutSessionStorageKey = 'iad.topology.layoutPositions.session.v2';

function loadLayoutPositions(): Record<string, { x: number; y: number }> {
  if (typeof window === 'undefined') return {};
  try {
    const raw =
      window.localStorage.getItem(layoutStorageKey) ??
      window.sessionStorage.getItem(layoutSessionStorageKey) ??
      (window as unknown as { __iadE2ELayout?: string }).__iadE2ELayout;
    if (!raw) return {};
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    return {};
  }
}

function saveLayoutPositions(value: Record<string, { x: number; y: number }>) {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(layoutStorageKey, JSON.stringify(value));
    window.sessionStorage.setItem(layoutSessionStorageKey, JSON.stringify(value));
  } catch {
    // Layout persistence is best-effort only.
  }
}

interface UIState {
  activeScreen: ScreenId;
  selectedDeviceId: string | null;
  selectedNodeId: string | null; // Added for TopologyScreen
  selectedEdgeId: string | null; // Added for TopologyScreen
  isSidebarOpen: boolean;
  isInspectorOpen: boolean;
  isLogsOpen: boolean;
  isSearchOpen: boolean;
  searchQuery: string;
  theme: 'light' | 'dark' | 'system';
  settings: { layoutEngine: LayoutEngine; packetAnimation: boolean; packetIntensity: PacketIntensity };
  layoutPositions: Record<string, { x: number; y: number }>;

  // Actions
  setActiveScreen: (screen: ScreenId) => void;
  setSelectedDeviceId: (id: string | null) => void;
  selectNode: (id: string | null) => void;
  selectEdge: (id: string | null) => void;
  setNodePosition: (id: string, position: { x: number; y: number }) => void;
  resetLayoutPositions: () => void;
  toggleSidebar: () => void;
  toggleInspector: () => void;
  toggleLogs: () => void;
  toggleSearch: () => void;
  setSearchQuery: (query: string) => void;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  setPacketAnimation: (enabled: boolean) => void;
  setPacketIntensity: (intensity: PacketIntensity) => void;
}

export const useUIStore = create<UIState>((set) => ({
  activeScreen: 'topology',
  selectedDeviceId: null,
  selectedNodeId: null,
  selectedEdgeId: null,
  isSidebarOpen: true,
  isInspectorOpen: false,
  isLogsOpen: true,
  isSearchOpen: false,
  searchQuery: '',
  theme: 'system',
  settings: { layoutEngine: 'elk_layered', packetAnimation: true, packetIntensity: 'normal' },
  layoutPositions: loadLayoutPositions(),

  setActiveScreen: (screen) => set({ activeScreen: screen }),

  setSelectedDeviceId: (id) => set(() => ({
      selectedDeviceId: id,
      isInspectorOpen: id !== null
  })),

  selectNode: (id) => set({ selectedNodeId: id, selectedEdgeId: null, selectedDeviceId: id, isInspectorOpen: id !== null }),
  selectEdge: (id) => set({ selectedEdgeId: id, selectedNodeId: null, isInspectorOpen: id !== null }),
  setNodePosition: (id, position) => set((state) => {
    const layoutPositions = { ...state.layoutPositions, [id]: position };
    saveLayoutPositions(layoutPositions);
    return { layoutPositions };
  }),
  resetLayoutPositions: () => {
    saveLayoutPositions({});
    if (typeof window !== 'undefined') {
      try {
        window.localStorage.removeItem(layoutStorageKey);
        window.sessionStorage.removeItem(layoutSessionStorageKey);
      } catch {
        // Ignore storage cleanup failures.
      }
    }
    set({ layoutPositions: {} });
  },

  toggleSidebar: () => set((state) => ({ isSidebarOpen: !state.isSidebarOpen })),
  toggleInspector: () => set((state) => ({ isInspectorOpen: !state.isInspectorOpen })),
  toggleLogs: () => set((state) => ({ isLogsOpen: !state.isLogsOpen })),
  toggleSearch: () => set((state) => ({ isSearchOpen: !state.isSearchOpen })),
  setSearchQuery: (query) => set({ searchQuery: query }),
  setTheme: (theme) => set({ theme }),
  setPacketAnimation: (enabled) => set((state) => ({ settings: { ...state.settings, packetAnimation: enabled } })),
  setPacketIntensity: (intensity) => set((state) => ({ settings: { ...state.settings, packetIntensity: intensity } })),
}));

if (typeof window !== 'undefined' && import.meta.env.DEV) {
  (window as unknown as {
    __iadTopologyLayout?: {
      setNodePosition: (id: string, position: { x: number; y: number }) => void;
      getLayoutPositions: () => Record<string, { x: number; y: number }>;
    };
  }).__iadTopologyLayout = {
    setNodePosition: (id, position) => useUIStore.getState().setNodePosition(id, position),
    getLayoutPositions: () => useUIStore.getState().layoutPositions,
  };
}
