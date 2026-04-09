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
    gap: 0.42rem;
    min-height: 2rem;
    padding: 0.38rem 0.72rem;
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
    padding: 0.28rem 0.58rem;
  }

  .token-badge__main {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    min-width: 0;
  }

  .token-badge__icon,
  .token-badge__count,
  .token-badge__trailing {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.28rem;
    height: 1.28rem;
    padding: 0 0.24rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--token-accent) 14%, var(--surface-raised));
    color: color-mix(in srgb, var(--token-accent) 86%, var(--text-primary));
    font-size: 0.6rem;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    flex: 0 0 auto;
  }

  .token-badge__label {
    min-width: 0;
    font-size: 0.74rem;
    font-weight: 650;
    letter-spacing: 0.01em;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .token-badge__count {
    min-width: 1.36rem;
    font-size: 0.67rem;
    box-shadow: inset 0 0 0 1px var(--border-soft);
  }
</style>
