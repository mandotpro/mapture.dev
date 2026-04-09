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
        <path d="M5 12h14"></path>
        <path d="M8 8l-4 4 4 4"></path>
        <path d="M16 8l4 4-4 4"></path>
      </svg>
    </span>
  {/snippet}
  {#snippet stamp()}
    <span></span><span></span>
  {/snippet}
</NodeFrame>
