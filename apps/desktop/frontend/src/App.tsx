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
import { useUIStore } from './store/useUIStore';
import { useScanStore } from './store/useScanStore';

export default function App() {
  const { activeScreen, setActiveScreen, isSidebarOpen, isInspectorOpen, isLogsOpen } = useUIStore();
  const hasReport = useScanStore((s) => s.report !== null);

  // Handle system dark mode matching
  useEffect(() => {
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      document.documentElement.classList.add('dark');
    }
  }, []);

  // Before any scan/import, the entry point is the Wireshark-style launch screen.
  const mainCanvas = !hasReport ? (
    <LaunchScreen />
  ) : activeScreen === 'devices' ? (
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
