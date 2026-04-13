<script lang="ts">
  let {
    type = 'button',
    href = '',
    target = '',
    rel = '',
    tone = 'soft',
    compact = false,
    active = false,
    disabled = false,
    className = '',
    title = '',
    ariaLabel = '',
    onclick,
    children,
  }: {
    type?: 'button' | 'submit' | 'reset';
    href?: string;
    target?: string;
    rel?: string;
    tone?: 'soft' | 'ghost' | 'subtle';
    compact?: boolean;
    active?: boolean;
    disabled?: boolean;
    className?: string;
    title?: string;
    ariaLabel?: string;
    onclick?: (event: MouseEvent) => void;
    children?: () => unknown;
  } = $props();

  const classes = $derived(
    ['ui-action', `ui-action--${tone}`, compact ? 'is-compact' : '', active ? 'is-active' : '', className]
      .filter(Boolean)
      .join(' '),
  );
</script>

{#if href}
  <a
    class={classes}
    href={href}
    target={target || undefined}
    rel={rel || undefined}
    title={title || undefined}
    aria-label={ariaLabel || undefined}
  >
    {@render children?.()}
  </a>
{:else}
  <button
    type={type}
    class={classes}
    disabled={disabled}
    title={title || undefined}
    aria-label={ariaLabel || undefined}
    onclick={onclick}
  >
    {@render children?.()}
  </button>
{/if}

<style>
  .ui-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    min-height: 2rem;
    padding: 0.4rem 0.72rem;
    border-radius: 999px;
    border: 1px solid var(--border-soft);
    background: color-mix(in srgb, var(--surface-panel) 92%, transparent);
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 24%, transparent);
    color: var(--text-primary);
    text-decoration: none;
    white-space: nowrap;
    cursor: pointer;
    transition:
      transform var(--ui-transition-fast),
      border-color var(--ui-transition-fast),
      background var(--ui-transition-fast),
      color var(--ui-transition-fast),
      box-shadow var(--ui-transition-fast);
  }

  .ui-action:hover {
    transform: translateY(-1px);
    border-color: var(--interactive-border-hover);
    background: var(--interactive-bg-hover);
    box-shadow: var(--interactive-shadow-hover);
  }

  .ui-action:active,
  .ui-action.is-active {
    transform: translateY(0);
    border-color: var(--interactive-border-selected);
    background: var(--interactive-bg-selected);
    color: var(--interactive-text-selected);
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 12%, transparent);
  }

  .ui-action:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 var(--focus-ring-width) var(--focus-ring),
      var(--interactive-shadow-hover);
  }

  .ui-action:disabled {
    opacity: 0.54;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }

  .ui-action:disabled:hover {
    border-color: var(--border-soft);
    background: color-mix(in srgb, var(--surface-panel) 92%, transparent);
  }

  .ui-action.is-compact {
    min-height: 1.7rem;
    padding: 0.26rem 0.58rem;
  }

  .ui-action--ghost {
    background: transparent;
    box-shadow: none;
  }

  .ui-action--subtle {
    background: color-mix(in srgb, var(--surface-panel-soft) 88%, transparent);
    box-shadow: none;
  }
</style>
