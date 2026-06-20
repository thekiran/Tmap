import { createContext, useContext } from 'react';
import type { PacketIntensity } from '../../lib/packet-flow';

/**
 * Global packet-animation configuration, provided once by TopologyScreen and
 * consumed by every animated edge. Keeping it in context (rather than per-edge
 * data) means toggling animation or intensity does not rebuild the edge array —
 * edges just re-read the context once.
 */
export interface PacketAnimationConfig {
  enabled: boolean;
  intensity: PacketIntensity;
  /** Per-edge particle cap, lowered automatically for dense graphs. */
  maxParticles: number;
}

export const PacketAnimationContext = createContext<PacketAnimationConfig>({
  enabled: true,
  intensity: 'normal',
  maxParticles: 3,
});

export const usePacketAnimation = (): PacketAnimationConfig => useContext(PacketAnimationContext);
