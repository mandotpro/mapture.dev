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
        <path d="M12 4l7 8-7 8-7-8 7-8z"></path>
        <circle cx="12" cy="12" r="1.6"></circle>
      </svg>
    </span>
  {/snippet}
  {#snippet stamp()}
    <span></span><span></span><span></span><span></span>
  {/snippet}
</NodeFrame>
