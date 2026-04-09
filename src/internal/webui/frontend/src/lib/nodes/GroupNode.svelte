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
        <rect x="4" y="6" width="7" height="6" rx="2"></rect>
        <rect x="13" y="6" width="7" height="6" rx="2"></rect>
        <rect x="8.5" y="13" width="7" height="6" rx="2"></rect>
      </svg>
    </span>
  {/snippet}
  {#snippet stamp()}
    <span></span><span></span><span></span>
  {/snippet}
</NodeFrame>
