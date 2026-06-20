import { describe, expect, it } from 'vitest';
import { derivePacketFlow, packetVisual } from './packet-flow';
import type { TopologyEdge } from './models';

function edge(partial: Partial<TopologyEdge> & Record<string, unknown>): TopologyEdge {
  return {
    id: 'e1',
    source: 'a',
    target: 'b',
    type: 'same_subnet',
    certainty: 'inferred',
    tier: 'l2',
    confidence: 0.5,
    label: '',
    basis: '',
    boundary: null,
    thin: false,
    physical: false,
    inferred: true,
    lineStyle: 'dashed',
    layers: ['l2'],
    ...partial,
  } as TopologyEdge;
}

describe('derivePacketFlow', () => {
  it('animates confirmed ARP links as bidirectional ARP traffic', () => {
    const flow = derivePacketFlow(edge({ certainty: 'confirmed', type: 'arp_confirmed', confidence: 0.9 }));
    expect(flow.trafficState).not.toBe('idle');
    expect(flow.animated).toBe(true);
    expect(flow.direction).toBe('bidirectional');
    expect(flow.protocol).toBe('arp');
    expect(flow.confidence).toBe('confirmed');
  });

  it('keeps inferred links idle (no evidence ⇒ no packets)', () => {
    const flow = derivePacketFlow(edge({ certainty: 'inferred', type: 'same_subnet', confidence: 0.6 }));
    expect(flow.trafficState).toBe('idle');
    expect(flow.animated).toBe(false);
    expect(packetVisual(flow, { intensity: 'high', selected: false, maxParticles: 3 })).toBeNull();
  });

  it('routes hierarchical gateway links forward', () => {
    const flow = derivePacketFlow(edge({ certainty: 'confirmed', type: 'gateway_default', confidence: 0.8 }));
    expect(flow.direction).toBe('forward');
  });

  it('uses a neutral bidirectional shimmer for unattributed links', () => {
    const flow = derivePacketFlow(edge({ certainty: 'confirmed', type: 'mystery_link', confidence: 0.9 }));
    expect(flow.direction).toBe('unknown');
    const visual = packetVisual(flow, { intensity: 'high', selected: false, maxParticles: 3 });
    expect(visual?.bidirectional).toBe(true);
    expect((visual?.count ?? 0)).toBeLessThanOrEqual(2);
  });

  it('lets explicit backend fields override the derivation', () => {
    const flow = derivePacketFlow(edge({ trafficState: 'high', direction: 'reverse', protocol: 'dns', certainty: 'inferred' }));
    expect(flow.trafficState).toBe('high');
    expect(flow.direction).toBe('reverse');
    expect(flow.protocol).toBe('dns');
    const visual = packetVisual(flow, { intensity: 'normal', selected: false, maxParticles: 3 });
    expect(visual?.reverse).toBe(true);
  });
});

describe('packetVisual performance caps', () => {
  it('clamps particle count to maxParticles on dense graphs', () => {
    const flow = derivePacketFlow(edge({ certainty: 'confirmed', type: 'arp_confirmed', confidence: 0.95 }));
    const visual = packetVisual(flow, { intensity: 'high', selected: false, maxParticles: 1 });
    expect(visual?.count).toBe(1);
  });

  it('speeds up packets at higher intensity', () => {
    const flow = derivePacketFlow(edge({ certainty: 'confirmed', type: 'arp_confirmed', confidence: 0.95 }));
    const low = packetVisual(flow, { intensity: 'low', selected: false, maxParticles: 3 })!;
    const high = packetVisual(flow, { intensity: 'high', selected: false, maxParticles: 3 })!;
    expect(high.durationMs).toBeLessThan(low.durationMs);
  });
});
