import { useEffect } from 'react';
import { DesktopLayout } from './components/layout/DesktopLayout';
import { MenuBar } from './components/layout/MenuBar';
import { Sidebar } from './components/layout/Sidebar';
import { BottomLogs } from './components/layout/BottomLogs';
import { Inspector } from './components/layout/Inspector';
import { TopologyScreen } from './components/topology/TopologyScreen';
import { DevicesScreen } from './components/devices/DevicesScreen';
import { EvidenceScreen } from './components/evidence/EvidenceScreen';
import { LaunchScreen } from './components/launch/LaunchScreen';
import { LaunchChrome } from './components/launch/LaunchChrome';
import { useUIStore } from './store/useUIStore';
import { useScanStore } from './store/useScanStore';
import { useScanEvents } from './lib/useScanEvents';

export default function App() {
  const { activeScreen, setActiveScreen, isSidebarOpen, isInspectorOpen, isLogsOpen } = useUIStore();
  const hasReport = useScanStore((s) => s.report !== null);
  const scanStatus = useScanStore((s) => s.scanStatus);

  // Subscribe the store to backend scan lifecycle events (or polling fallback).
  useScanEvents();

  // Handle system dark mode matching
  useEffect(() => {
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      document.documentElement.classList.add('dark');
    }
  }, []);

  // The launch screen is only the entry point while truly idle. As soon as a
  // scan starts — or a report is loaded — we swap to the full workspace so the
  // live topology dashboard is visible and the UI never blocks on "Scanning…".
  const inWorkspace = hasReport || scanStatus !== 'idle';
  if (!inWorkspace) {
    return (
      <LaunchChrome>
        <LaunchScreen />
      </LaunchChrome>
    );
  }

  const mainCanvas =
    activeScreen === 'devices' ? (
      <DevicesScreen />
    ) : activeScreen === 'evidence' ? (
      <EvidenceScreen />
    ) : (
      <TopologyScreen />
    );

  return (
    <DesktopLayout
      topBar={<MenuBar />}
      sidebar={
        <Sidebar
          activeScreen={activeScreen}
          onScreenChange={(screen) => setActiveScreen(screen as never)}
        />
      }
      mainCanvas={mainCanvas}
      inspector={<Inspector />}
      bottomLogs={<BottomLogs />}
      isSidebarOpen={isSidebarOpen}
      isInspectorOpen={isInspectorOpen}
      isLogsOpen={isLogsOpen}
    />
  );
}
