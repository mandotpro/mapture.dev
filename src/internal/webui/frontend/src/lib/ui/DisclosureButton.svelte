<script lang="ts">
  let {
    icon = '',
    title = '',
    summary = '',
    open = false,
    className = '',
    onclick,
  }: {
    icon?: string;
    title: string;
    summary?: string;
    open?: boolean;
    className?: string;
    onclick?: (event: MouseEvent) => void;
  } = $props();
</script>

<button
  type="button"
  class={['ui-disclosure', open ? 'is-open' : '', className].filter(Boolean).join(' ')}
  aria-expanded={open}
  onclick={onclick}
>
  <span class="ui-disclosure__icon" aria-hidden="true">{icon}</span>
  <span class="ui-disclosure__copy">
    <strong>{title}</strong>
    {#if summary}
      <small>{summary}</small>
    {/if}
  </span>
  <span class={['ui-disclosure__caret', open ? 'is-open' : ''].join(' ')} aria-hidden="true">
    <svg viewBox="0 0 16 16" focusable="false">
      <path d="M4.5 6.25 8 9.75l3.5-3.5"></path>
    </svg>
  </span>
</button>

<style>
  .ui-disclosure {
    display: inline-flex;
    align-items: center;
    gap: 0.58rem;
    width: 100%;
    min-width: 224px;
    padding: 0.42rem 0.5rem;
    border-radius: 18px;
    border: 1px solid var(--border-soft);
    background: color-mix(in srgb, var(--surface-panel) 92%, transparent);
    cursor: pointer;
    transition:
      transform var(--ui-transition-fast),
      border-color var(--ui-transition-fast),
      background var(--ui-transition-fast),
      box-shadow var(--ui-transition-fast);
  }

  .ui-disclosure:hover {
    transform: translateY(-1px);
    border-color: var(--interactive-border-hover);
    background: var(--interactive-bg-hover);
    box-shadow: var(--interactive-shadow-hover);
  }

  .ui-disclosure.is-open {
    border-color: var(--interactive-border-selected);
    background: var(--interactive-bg-selected);
  }

  .ui-disclosure:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 var(--focus-ring-width) var(--focus-ring),
      var(--interactive-shadow-hover);
  }

  .ui-disclosure__icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.8rem;
    height: 1.8rem;
    border-radius: 14px;
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: 0.6rem;
    font-weight: 800;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--text-primary) 8%, transparent);
  }

  .ui-disclosure__copy {
    display: grid;
    justify-items: start;
    gap: 0.08rem;
    min-width: 0;
    flex: 1 1 auto;
  }

  .ui-disclosure__copy strong {
    color: var(--text-primary);
    font-size: 0.75rem;
  }

  .ui-disclosure__copy small {
    color: var(--text-secondary);
    font-size: 0.66rem;
  }

  .ui-disclosure__caret {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.15rem;
    height: 1.15rem;
    color: var(--text-secondary);
    transition: transform var(--ui-transition-fast), color var(--ui-transition-fast);
  }

  .ui-disclosure__caret svg {
    width: 0.9rem;
    height: 0.9rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.7;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .ui-disclosure__caret.is-open {
    transform: rotate(180deg);
    color: var(--text-primary);
  }
</style>
