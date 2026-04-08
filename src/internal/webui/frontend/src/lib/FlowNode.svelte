<script lang="ts">
  import { Handle, Position, type NodeProps } from '@xyflow/svelte';

  let {
    data,
    selected = false,
    sourcePosition = Position.Right,
    targetPosition = Position.Left,
  }: NodeProps = $props();
</script>

<Handle type="target" position={targetPosition} />
<article class={['mapture-node', data.type, selected ? 'selected' : ''].join(' ')}>
  <div class="mapture-node__eyebrow">
    <span class="mapture-node__dot"></span>
    <span>{data.type}</span>
  </div>
  <strong>{data.label}</strong>
  {#if data.subtitle}
    <p>{data.subtitle}</p>
  {/if}
</article>
<Handle type="source" position={sourcePosition} />
