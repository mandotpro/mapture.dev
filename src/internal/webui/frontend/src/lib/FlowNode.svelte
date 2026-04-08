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
  class={['mapture-node', `mapture-node--${data.type}`, selected ? 'selected' : ''].join(' ')}
  style={`--node-color:${data.color};`}
>
  <div class="mapture-node__shell">
    <header class="mapture-node__header">
      <div class="mapture-node__eyebrow">
        <span class="mapture-node__glyph" aria-hidden="true">
          {#if data.type === 'service'}
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
        <span>{typeLabel(data.type)}</span>
      </div>

      <span class="mapture-node__stamp" aria-hidden="true">
        {#if data.type === 'service'}
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
  </div>
</article>
<Handle type="source" position={sourcePosition} />
