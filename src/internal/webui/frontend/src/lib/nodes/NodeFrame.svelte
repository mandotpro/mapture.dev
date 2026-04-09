<script lang="ts">
  import { Handle, Position, type NodeProps } from '@xyflow/svelte';
  import type { ArchitectureNodeData } from '../types';

  let {
    data,
    selected = false,
    sourcePosition = Position.Right,
    targetPosition = Position.Left,
    glyph,
    stamp,
  }: NodeProps<ArchitectureNodeData> & {
    glyph?: () => unknown;
    stamp?: () => unknown;
  } = $props();

  function typeLabel(type: string): string {
    return type.replaceAll('_', ' ');
  }
</script>

<Handle type="target" position={targetPosition} />
<article
  class={[
    'mapture-node',
    `mapture-node--${data.type}`,
    `mapture-node--kind-${data.kind ?? 'node'}`,
    data.groupKind ? `mapture-node--group-${data.groupKind}` : '',
    data.trace ? 'mapture-node--trace' : '',
    data.impact && data.impact !== 'none' ? `mapture-node--impact-${data.impact}` : '',
    `mapture-node--tone-${data.tone ?? 'primary'}`,
    `mapture-node--mode-${data.viewMode ?? 'system-map'}`,
    selected ? 'selected' : '',
  ].join(' ')}
  style={`--node-color:${data.color};`}
>
  <div class="mapture-node__shell">
    <header class="mapture-node__header">
      <div class="mapture-node__eyebrow">
        {@render glyph?.()}
        <span>{data.eyebrow ?? typeLabel(data.type)}</span>
      </div>

      <span class="mapture-node__stamp" aria-hidden="true">
        {@render stamp?.()}
      </span>
    </header>

    <strong>{data.label}</strong>

    {#if data.subtitle}
      <p>{data.subtitle}</p>
    {/if}

    {#if data.kind !== 'node'}
      <span class="mapture-node__metric">{data.memberCount} nodes</span>
    {/if}
  </div>
</article>
<Handle type="source" position={sourcePosition} />
