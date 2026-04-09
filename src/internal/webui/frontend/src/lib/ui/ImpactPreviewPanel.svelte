<script lang="ts">
  import type { ImpactPreview, PresentedNode } from '../types';
  import TokenBadge from './TokenBadge.svelte';

  let {
    preview,
    ontrace,
  }: {
    preview: ImpactPreview;
    ontrace?: (slot: 'source' | 'target', node: PresentedNode) => void;
  } = $props();
</script>

<article class="impact-panel" data-interactive-root>
  <header class="impact-panel__head">
    <div>
      <strong>Impact Preview</strong>
      <p>Quick upstream and downstream reach for the selected node.</p>
    </div>
    {#if preview.crossBoundaryTouches > 0}
      <TokenBadge
        label="Boundary"
        count={preview.crossBoundaryTouches}
        accent="var(--warning)"
        interactive={false}
        compact
        quiet
      />
    {/if}
  </header>

  <div class="impact-panel__metrics">
    <section class="impact-panel__metric">
      <span>Downstream</span>
      <strong>{preview.directDownstream.length}</strong>
      <small>{preview.downstreamReach} reachable</small>
    </section>
    <section class="impact-panel__metric">
      <span>Upstream</span>
      <strong>{preview.directUpstream.length}</strong>
      <small>{preview.upstreamReach} reachable</small>
    </section>
  </div>

  {#if preview.directDownstream.length > 0}
    <section class="impact-panel__list">
      <span>Immediate downstream</span>
      <div class="impact-panel__chips">
        {#each preview.directDownstream.slice(0, 4) as node}
          <TokenBadge
            label={node.name}
            accent={node.colorHint || 'var(--accent)'}
            interactive={Boolean(ontrace)}
            compact
            onclick={ontrace ? () => ontrace('target', node) : undefined}
          />
        {/each}
      </div>
    </section>
  {/if}

  {#if preview.directUpstream.length > 0}
    <section class="impact-panel__list">
      <span>Immediate upstream</span>
      <div class="impact-panel__chips">
        {#each preview.directUpstream.slice(0, 4) as node}
          <TokenBadge
            label={node.name}
            accent={node.colorHint || 'var(--accent)'}
            interactive={Boolean(ontrace)}
            compact
            onclick={ontrace ? () => ontrace('source', node) : undefined}
          />
        {/each}
      </div>
    </section>
  {/if}
</article>

<style>
  .impact-panel {
    pointer-events: auto;
    width: min(360px, calc(100vw - 1.5rem));
    display: grid;
    gap: 0.72rem;
    padding: 0.88rem;
    border-radius: 24px;
    border: 1px solid var(--border-soft);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-floating);
    backdrop-filter: blur(16px);
  }

  .impact-panel__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.8rem;
  }

  .impact-panel__head strong {
    display: block;
    color: var(--text-primary);
    font-size: 0.82rem;
  }

  .impact-panel__head p {
    margin: 0.18rem 0 0;
    color: var(--text-secondary);
    font-size: 0.72rem;
    line-height: 1.45;
  }

  .impact-panel__metrics {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.45rem;
  }

  .impact-panel__metric {
    display: grid;
    gap: 0.08rem;
    padding: 0.66rem;
    border-radius: 18px;
    border: 1px solid var(--border-soft);
    background: var(--surface-panel-soft);
  }

  .impact-panel__metric span {
    color: var(--text-tertiary);
    font-size: 0.66rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .impact-panel__metric strong {
    color: var(--text-primary);
    font-size: 1rem;
  }

  .impact-panel__metric small {
    color: var(--text-secondary);
    font-size: 0.7rem;
  }

  .impact-panel__list {
    display: grid;
    gap: 0.32rem;
  }

  .impact-panel__list > span {
    color: var(--text-tertiary);
    font-size: 0.66rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .impact-panel__chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.34rem;
  }
</style>
