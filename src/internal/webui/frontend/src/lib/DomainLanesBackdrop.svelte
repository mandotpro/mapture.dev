<svelte:options runes={true} />

<script lang="ts">
  import { ViewportPortal } from '@xyflow/svelte';
  import type { LaneOverlay } from './types';

  let { lanes = [] }: { lanes?: LaneOverlay[] } = $props();
</script>

{#if lanes.length > 0}
  <ViewportPortal target="back">
    <div class="domain-lanes-backdrop" aria-hidden="true">
      {#each lanes as lane}
        <section
          class="domain-lane"
          style={`left:${lane.x}px;top:${lane.top}px;width:${lane.width}px;height:${lane.height}px;--lane-accent:${lane.accent};`}
        >
          <header class="domain-lane__head">
            <strong>{lane.label}</strong>
            {#if lane.ownerLabel}
              <small>{lane.ownerLabel}</small>
            {/if}
          </header>
        </section>
      {/each}
    </div>
  </ViewportPortal>
{/if}
