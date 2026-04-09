<script lang="ts">
  import { Handle, Position, type NodeProps } from '@xyflow/svelte';

  let {
    data,
    selected = false,
    sourcePosition = Position.Right,
    targetPosition = Position.Left,
  }: NodeProps = $props();

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
        <span class="mapture-node__glyph" aria-hidden="true">
          {#if data.kind === 'group'}
            <svg viewBox="0 0 24 24" focusable="false">
              <rect x="4" y="6" width="7" height="6" rx="2"></rect>
              <rect x="13" y="6" width="7" height="6" rx="2"></rect>
              <rect x="8.5" y="13" width="7" height="6" rx="2"></rect>
            </svg>
          {:else if data.kind === 'bridge'}
            <svg viewBox="0 0 24 24" focusable="false">
              <path d="M5 12h14"></path>
              <path d="M8 8l-4 4 4 4"></path>
              <path d="M16 8l4 4-4 4"></path>
            </svg>
          {:else if data.type === 'service'}
            <svg viewBox="0 0 24 24" focusable="false">
              <rect x="5" y="6" width="14" height="12" rx="3"></rect>
              <path d="M9 3v3M15 3v3M9 18v3M15 18v3M3 9h3M18 9h3M3 15h3M18 15h3"></path>
            </svg>
          {:else if data.type === 'api'}
            <svg viewBox="0 0 24 24" focusable="false">
              <path d="M5 7h14M5 12h14M5 17h14"></path>
              <circle cx="8" cy="7" r="1.4"></circle>
              <circle cx="12" cy="12" r="1.4"></circle>
              <circle cx="16" cy="17" r="1.4"></circle>
            </svg>
          {:else if data.type === 'database'}
            <svg viewBox="0 0 24 24" focusable="false">
              <ellipse cx="12" cy="6.5" rx="6.5" ry="3"></ellipse>
              <path d="M5.5 6.5v8c0 1.7 2.9 3 6.5 3s6.5-1.3 6.5-3v-8"></path>
              <path d="M5.5 10.5c0 1.7 2.9 3 6.5 3s6.5-1.3 6.5-3"></path>
            </svg>
          {:else}
            <svg viewBox="0 0 24 24" focusable="false">
              <path d="M12 4l7 8-7 8-7-8 7-8z"></path>
              <circle cx="12" cy="12" r="1.6"></circle>
            </svg>
          {/if}
        </span>
        <span>{data.eyebrow ?? typeLabel(data.type)}</span>
      </div>

      <span class="mapture-node__stamp" aria-hidden="true">
        {#if data.kind === 'group'}
          <span></span><span></span><span></span>
        {:else if data.kind === 'bridge'}
          <span></span><span></span>
        {:else if data.type === 'service'}
          <span></span><span></span><span></span>
        {:else if data.type === 'api'}
          <span></span><span></span>
        {:else if data.type === 'database'}
          <span></span>
        {:else}
          <span></span><span></span><span></span><span></span>
        {/if}
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
