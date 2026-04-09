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
        <path d="M5 7h14M5 12h14M5 17h14"></path>
        <circle cx="8" cy="7" r="1.4"></circle>
        <circle cx="12" cy="12" r="1.4"></circle>
        <circle cx="16" cy="17" r="1.4"></circle>
      </svg>
    </span>
  {/snippet}
  {#snippet stamp()}
    <span></span><span></span>
  {/snippet}
</NodeFrame>
