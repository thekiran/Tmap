import { create } from 'zustand';
import type { LayoutEngine } from '../lib/models';

export type ScreenId = 'overview' | 'topology' | 'devices' | 'evidence' | 'probes' | 'export';

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
  settings: { layoutEngine: LayoutEngine };
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
  settings: { layoutEngine: 'elk_layered' },
  layoutPositions: {},

  setActiveScreen: (screen) => set({ activeScreen: screen }),

  setSelectedDeviceId: (id) => set(() => ({
      selectedDeviceId: id,
      isInspectorOpen: id !== null
  })),

  selectNode: (id) => set({ selectedNodeId: id, selectedEdgeId: null, selectedDeviceId: id, isInspectorOpen: id !== null }),
  selectEdge: (id) => set({ selectedEdgeId: id, selectedNodeId: null, isInspectorOpen: id !== null }),
  setNodePosition: (id, position) => set((state) => ({ layoutPositions: { ...state.layoutPositions, [id]: position } })),
  resetLayoutPositions: () => set({ layoutPositions: {} }),

  toggleSidebar: () => set((state) => ({ isSidebarOpen: !state.isSidebarOpen })),
  toggleInspector: () => set((state) => ({ isInspectorOpen: !state.isInspectorOpen })),
  toggleLogs: () => set((state) => ({ isLogsOpen: !state.isLogsOpen })),
  toggleSearch: () => set((state) => ({ isSearchOpen: !state.isSearchOpen })),
  setSearchQuery: (query) => set({ searchQuery: query }),
  setTheme: (theme) => set({ theme }),
}));
