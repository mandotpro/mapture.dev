<script lang="ts">
  import type { NodeInspectorAction, PresentedNode } from '../types';
  import TokenBadge from './TokenBadge.svelte';

  let {
    node,
    badgeLabel,
    badgeAccent,
    domainLabel,
    ownerLabel,
    sourceLabel,
    compositionLabel = '',
    summary = '',
    actions = [],
    onaction,
    onclose,
  }: {
    node: PresentedNode;
    badgeLabel: string;
    badgeAccent: string;
    domainLabel: string;
    ownerLabel: string;
    sourceLabel: string;
    compositionLabel?: string;
    summary?: string;
    actions?: NodeInspectorAction[];
    onaction?: (id: string) => void;
    onclose?: () => void;
  } = $props();

  function trigger(actionId: string): void {
    onaction?.(actionId);
  }
</script>

<article class="node-inspector" data-interactive-root>
  <header class="node-inspector__head">
    <div class="node-inspector__title">
      <TokenBadge
        label={badgeLabel}
        accent={badgeAccent}
        interactive={false}
        compact
        quiet
        className="node-inspector__badge"
      />
      <strong>{node.name}</strong>
      <small>{node.id}</small>
    </div>

    <button type="button" class="node-inspector__close" aria-label="Close node overview" onclick={onclose}>
      <span aria-hidden="true">x</span>
    </button>
  </header>

  <div class="node-inspector__grid">
    <section class="node-inspector__meta">
      <span>Domain</span>
      <strong>{domainLabel}</strong>
    </section>
    <section class="node-inspector__meta">
      <span>Owner</span>
      <strong>{ownerLabel}</strong>
    </section>
    <section class="node-inspector__meta">
      <span>{node.kind === 'node' ? 'Source' : 'Composition'}</span>
      <strong>{node.kind === 'node' ? sourceLabel : compositionLabel || 'n/a'}</strong>
    </section>
  </div>

  {#if summary}
    <p class="node-inspector__summary">{summary}</p>
  {/if}

  {#if actions.length > 0}
    <div class="node-inspector__actions">
      {#each actions as action}
        <TokenBadge
          label={action.label}
          accent={action.tone === 'accent' ? badgeAccent : 'var(--accent)'}
          interactive
          compact
          trailingText={action.badge ?? ''}
          className="node-inspector__action"
          onclick={() => trigger(action.id)}
        />
      {/each}
    </div>
  {/if}
</article>

<style>
  .node-inspector {
    pointer-events: auto;
    width: min(360px, calc(100vw - 1.5rem));
    display: grid;
    gap: 0.84rem;
    padding: 0.94rem;
    border-radius: 26px;
    border: 1px solid var(--border-strong);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 10%, transparent), transparent 36%),
      var(--surface-overlay);
    box-shadow: var(--shadow-floating);
    backdrop-filter: blur(18px);
  }

  .node-inspector__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.9rem;
  }

  .node-inspector__title {
    display: grid;
    gap: 0.18rem;
    min-width: 0;
  }

  .node-inspector__title strong {
    font-family: "Iowan Old Style", "Palatino Linotype", serif;
    font-size: 1.08rem;
    line-height: 1.2;
    color: var(--text-primary);
  }

  .node-inspector__title small {
    color: var(--text-secondary);
    font-size: 0.72rem;
    font-family: "SFMono-Regular", "Consolas", monospace;
    overflow-wrap: anywhere;
  }

  .node-inspector__close {
    width: 2.1rem;
    height: 2.1rem;
    padding: 0;
    border-radius: 999px;
    box-shadow: none;
    background: var(--surface-panel);
  }

  .node-inspector__close span {
    font-size: 0.74rem;
    font-weight: 800;
    text-transform: uppercase;
  }

  .node-inspector__grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.5rem;
  }

  .node-inspector__meta {
    display: grid;
    gap: 0.16rem;
    padding: 0.66rem 0.7rem;
    border-radius: 18px;
    border: 1px solid var(--border-soft);
    background: var(--surface-panel-soft);
  }

  .node-inspector__meta span {
    color: var(--text-tertiary);
    font-size: 0.64rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .node-inspector__meta strong {
    color: var(--text-primary);
    font-size: 0.76rem;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .node-inspector__summary {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.8rem;
    line-height: 1.5;
  }

  .node-inspector__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.42rem;
  }

  @media (max-width: 760px) {
    .node-inspector {
      width: min(100vw - 1rem, 360px);
    }

    .node-inspector__grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
