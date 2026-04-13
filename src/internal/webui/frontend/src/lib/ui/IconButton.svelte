<script lang="ts">
  let {
    href = '',
    target = '',
    rel = '',
    active = false,
    subtle = false,
    disabled = false,
    className = '',
    title = '',
    ariaLabel = '',
    onclick,
    children,
  }: {
    href?: string;
    target?: string;
    rel?: string;
    active?: boolean;
    subtle?: boolean;
    disabled?: boolean;
    className?: string;
    title?: string;
    ariaLabel?: string;
    onclick?: (event: MouseEvent) => void;
    children?: () => unknown;
  } = $props();

  const classes = $derived(
    ['ui-icon-button', subtle ? 'is-subtle' : '', active ? 'is-active' : '', className].filter(Boolean).join(' '),
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
    type="button"
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
  .ui-icon-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.15rem;
    height: 2.15rem;
    padding: 0;
    border-radius: 999px;
    border: 1px solid transparent;
    background: color-mix(in srgb, var(--surface-panel) 92%, transparent);
    color: var(--text-primary);
    text-decoration: none;
    cursor: pointer;
    transition:
      transform var(--ui-transition-fast),
      border-color var(--ui-transition-fast),
      background var(--ui-transition-fast),
      color var(--ui-transition-fast),
      box-shadow var(--ui-transition-fast);
  }

  .ui-icon-button:hover {
    transform: translateY(-1px);
    border-color: var(--interactive-border-hover);
    background: var(--interactive-bg-hover);
    box-shadow: var(--interactive-shadow-hover);
  }

  .ui-icon-button.is-active {
    border-color: var(--interactive-border-selected);
    background: var(--interactive-bg-selected);
    color: var(--interactive-text-selected);
  }

  .ui-icon-button.is-subtle {
    background: transparent;
    box-shadow: none;
  }

  .ui-icon-button:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 var(--focus-ring-width) var(--focus-ring),
      var(--interactive-shadow-hover);
  }

  .ui-icon-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }
</style>
