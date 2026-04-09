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
        <rect x="5" y="6" width="14" height="12" rx="3"></rect>
        <path d="M9 3v3M15 3v3M9 18v3M15 18v3M3 9h3M18 9h3M3 15h3M18 15h3"></path>
      </svg>
    </span>
  {/snippet}
  {#snippet stamp()}
    <span></span><span></span><span></span>
  {/snippet}
</NodeFrame>
