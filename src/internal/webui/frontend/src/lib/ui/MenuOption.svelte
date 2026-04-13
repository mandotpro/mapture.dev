<script lang="ts">
  let {
    icon = '',
    title = '',
    description = '',
    active = false,
    disabled = false,
    className = '',
    onclick,
  }: {
    icon?: string;
    title: string;
    description?: string;
    active?: boolean;
    disabled?: boolean;
    className?: string;
    onclick?: (event: MouseEvent) => void;
  } = $props();
</script>

<button
  type="button"
  class={['ui-menu-option', active ? 'is-active' : '', className].filter(Boolean).join(' ')}
  disabled={disabled}
  onclick={onclick}
>
  <span class="ui-menu-option__icon" aria-hidden="true">{icon}</span>
  <span class="ui-menu-option__copy">
    <strong>{title}</strong>
    {#if description}
      <small>{description}</small>
    {/if}
  </span>
</button>

<style>
  .ui-menu-option {
    display: inline-flex;
    align-items: center;
    justify-content: flex-start;
    gap: 0.58rem;
    width: 100%;
    padding: 0.38rem 0.44rem;
    border-radius: 18px;
    border: 1px solid transparent;
    background: transparent;
    box-shadow: none;
    cursor: pointer;
    transition:
      transform var(--ui-transition-fast),
      border-color var(--ui-transition-fast),
      background var(--ui-transition-fast),
      box-shadow var(--ui-transition-fast);
  }

  .ui-menu-option:hover {
    transform: translateY(-1px);
    border-color: var(--interactive-border-hover);
    background: color-mix(in srgb, var(--surface-panel-soft) 94%, transparent);
  }

  .ui-menu-option.is-active {
    border-color: var(--interactive-border-selected);
    background: var(--interactive-bg-selected);
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 10%, transparent);
  }

  .ui-menu-option:focus-visible {
    outline: none;
    box-shadow:
      0 0 0 var(--focus-ring-width) var(--focus-ring),
      inset 0 0 0 1px color-mix(in srgb, var(--accent) 10%, transparent);
  }

  .ui-menu-option:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }

  .ui-menu-option__icon {
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

  .ui-menu-option__copy {
    display: grid;
    justify-items: start;
    gap: 0.08rem;
    min-width: 0;
    flex: 1 1 auto;
  }

  .ui-menu-option__copy strong {
    color: var(--text-primary);
    font-size: 0.75rem;
  }

  .ui-menu-option__copy small {
    color: var(--text-secondary);
    font-size: 0.66rem;
  }
</style>
