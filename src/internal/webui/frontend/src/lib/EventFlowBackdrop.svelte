<svelte:options runes={true} />

<script lang="ts">
  import { ViewportPortal } from '@xyflow/svelte';
  import type { StageBandOverlay } from './types';

  let { bands = [] }: { bands?: StageBandOverlay[] } = $props();
</script>

{#if bands.length > 0}
  <ViewportPortal target="back">
    <div class="event-flow-backdrop" aria-hidden="true">
      {#each bands as band}
        <section
          class="event-flow-band"
          style={`left:${band.x}px;top:${band.top}px;width:${band.width}px;height:${band.height}px;--band-accent:${band.accent};`}
        >
          <header class="event-flow-band__head">
            <strong>{band.label}</strong>
            <small>{band.summary}</small>
          </header>
        </section>
      {/each}
    </div>
  </ViewportPortal>
{/if}
