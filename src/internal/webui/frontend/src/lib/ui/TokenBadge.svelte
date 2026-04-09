<script lang="ts">
  let {
    label = '',
    icon = '',
    count = null,
    accent = 'var(--accent)',
    active = false,
    compact = false,
    quiet = false,
    interactive = true,
    trailingText = '',
    className = '',
    style = '',
    title = '',
    ariaLabel = '',
    onclick,
  }: {
    label?: string;
    icon?: string;
    count?: number | string | null;
    accent?: string;
    active?: boolean;
    compact?: boolean;
    quiet?: boolean;
    interactive?: boolean;
    trailingText?: string;
    className?: string;
    style?: string;
    title?: string;
    ariaLabel?: string;
    onclick?: (event: MouseEvent) => void;
  } = $props();

  const badgeClass = $derived(
    [
      'token-badge',
      active ? 'is-active' : '',
      compact ? 'is-compact' : '',
      quiet ? 'is-quiet' : '',
      className,
    ]
      .filter(Boolean)
      .join(' '),
  );
  const badgeStyle = $derived(`--token-accent:${accent};${style}`);
</script>

{#if interactive}
  <button
    type="button"
    class={badgeClass}
    style={badgeStyle}
    title={title || undefined}
    aria-label={ariaLabel || undefined}
    onclick={onclick}
  >
    <span class="token-badge__main">
      {#if icon}
        <span class="token-badge__icon" aria-hidden="true">{icon}</span>
      {/if}
      {#if label}
        <span class="token-badge__label">{label}</span>
      {/if}
    </span>

    {#if count !== null && count !== undefined && count !== ''}
      <span class="token-badge__count">{count}</span>
    {/if}

    {#if trailingText}
      <span class="token-badge__trailing" aria-hidden="true">{trailingText}</span>
    {/if}
  </button>
{:else}
  <span
    class={badgeClass}
    style={badgeStyle}
    title={title || undefined}
    aria-label={ariaLabel || undefined}
  >
    <span class="token-badge__main">
      {#if icon}
        <span class="token-badge__icon" aria-hidden="true">{icon}</span>
      {/if}
      {#if label}
        <span class="token-badge__label">{label}</span>
      {/if}
    </span>

    {#if count !== null && count !== undefined && count !== ''}
      <span class="token-badge__count">{count}</span>
    {/if}

    {#if trailingText}
      <span class="token-badge__trailing" aria-hidden="true">{trailingText}</span>
    {/if}
  </span>
{/if}

<style>
  .token-badge {
    --token-accent: var(--accent);
    display: inline-flex;
    align-items: center;
    gap: 0.34rem;
    min-height: 2rem;
    padding: 0.34rem 0.66rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--token-accent) 18%, var(--border-soft));
    background: color-mix(in srgb, var(--token-accent) 8%, var(--surface-raised));
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 32%, transparent);
    color: var(--text-primary);
    white-space: nowrap;
    transition: transform 140ms ease, border-color 140ms ease, background 140ms ease, box-shadow 140ms ease;
  }

  .token-badge:hover {
    transform: translateY(-1px);
    border-color: color-mix(in srgb, var(--token-accent) 28%, var(--border-strong));
  }

  .token-badge.is-active {
    background: color-mix(in srgb, var(--token-accent) 14%, var(--surface-panel));
    border-color: color-mix(in srgb, var(--token-accent) 36%, var(--border-strong));
    color: color-mix(in srgb, var(--token-accent) 78%, var(--text-primary));
  }

  .token-badge.is-quiet {
    background: var(--surface-panel-soft);
  }

  .token-badge.is-compact {
    min-height: 1.8rem;
    padding: 0.24rem 0.52rem;
  }

  .token-badge__main {
    display: inline-flex;
    align-items: center;
    gap: 0.34rem;
    min-width: 0;
  }

  .token-badge__icon,
  .token-badge__count,
  .token-badge__trailing {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.08rem;
    height: 1.08rem;
    padding: 0 0.22rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--token-accent) 10%, var(--surface-raised));
    color: color-mix(in srgb, var(--token-accent) 74%, var(--text-primary));
    font-size: 0.55rem;
    font-weight: 800;
    letter-spacing: 0.03em;
    text-transform: uppercase;
    flex: 0 0 auto;
  }

  .token-badge__label {
    min-width: 0;
    font-size: 0.72rem;
    font-weight: 620;
    letter-spacing: 0.01em;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .token-badge__count {
    min-width: 1.18rem;
    height: 1rem;
    padding: 0 0.22rem;
    font-size: 0.62rem;
    box-shadow: inset 0 0 0 1px var(--border-soft);
  }

  .token-badge__trailing {
    min-width: 1rem;
    height: 1rem;
    font-size: 0.56rem;
  }
</style>
