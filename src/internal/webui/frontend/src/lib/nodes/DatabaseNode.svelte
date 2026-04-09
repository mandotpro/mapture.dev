<script lang="ts">
  import type { NodeProps } from '@xyflow/svelte';
  import type { ArchitectureNodeData } from '../types';
  import NodeFrame from './NodeFrame.svelte';

  let props: NodeProps<ArchitectureNodeData> = $props();
</script>

<NodeFrame {...props}>
  {#snippet glyph()}
    <span class="mapture-node__glyph" aria-hidden="true">
      <svg viewBox="0 0 24 24" focusable="false">
        <ellipse cx="12" cy="6.5" rx="6.5" ry="3"></ellipse>
        <path d="M5.5 6.5v8c0 1.7 2.9 3 6.5 3s6.5-1.3 6.5-3v-8"></path>
        <path d="M5.5 10.5c0 1.7 2.9 3 6.5 3s6.5-1.3 6.5-3"></path>
      </svg>
    </span>
  {/snippet}
  {#snippet stamp()}
    <span></span>
  {/snippet}
</NodeFrame>
