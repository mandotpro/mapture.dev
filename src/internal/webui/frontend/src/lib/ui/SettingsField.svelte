<script lang="ts">
  import type { SettingsFieldConfig } from '../types';

  let {
    field,
    onchange,
  }: {
    field: SettingsFieldConfig;
    onchange?: (id: string, value: boolean | string) => void;
  } = $props();

  function updateBoolean(next: boolean): void {
    onchange?.(field.id, next);
  }

  function updateString(next: string): void {
    onchange?.(field.id, next);
  }
</script>

<article class={['settings-field', field.disabled ? 'is-disabled' : ''].join(' ')}>
  <div class="settings-field__copy">
    <div class="settings-field__label-row">
      <strong>{field.label}</strong>
      {#if field.badge}
        <span class="settings-field__badge">{field.badge}</span>
      {/if}
    </div>
    <p>{field.description}</p>
  </div>

  <div class="settings-field__action">
    {#if field.kind === 'toggle'}
      <button
        type="button"
        class={['settings-toggle', field.value ? 'is-active' : ''].join(' ')}
        onclick={() => updateBoolean(!field.value)}
        disabled={field.disabled}
      >
        <span class="settings-toggle__track">
          <span class="settings-toggle__thumb"></span>
        </span>
        <small>{field.value ? 'On' : 'Off'}</small>
      </button>
    {:else if field.kind === 'checkbox'}
      <label class="settings-checkbox">
        <input
          type="checkbox"
          checked={field.value}
          onchange={(event) => updateBoolean((event.currentTarget as HTMLInputElement).checked)}
          disabled={field.disabled}
        />
        <span>{field.value ? 'Enabled' : 'Disabled'}</span>
      </label>
    {:else if field.kind === 'choice'}
      <div class="settings-choice" role="group" aria-label={field.label}>
        {#each field.options as option}
          <button
            type="button"
            class={['settings-choice__option', field.value === option.value ? 'is-active' : ''].join(' ')}
            onclick={() => updateString(option.value)}
            disabled={field.disabled}
            title={option.description}
          >
            {#if option.glyph}
              <span class="settings-choice__glyph" aria-hidden="true">{option.glyph}</span>
            {/if}
            <span>{option.label}</span>
          </button>
        {/each}
      </div>
    {:else}
      <input
        class="settings-input"
        type={field.inputType ?? 'text'}
        value={field.value}
        placeholder={field.placeholder}
        oninput={(event) => updateString((event.currentTarget as HTMLInputElement).value)}
        disabled={field.disabled}
      />
    {/if}
  </div>
</article>

<style>
  .settings-field {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.88rem 0.92rem;
    border-radius: 20px;
    border: 1px solid var(--border-soft);
    background: var(--surface-raised);
  }

  .settings-field.is-disabled {
    opacity: 0.58;
  }

  .settings-field__copy {
    display: grid;
    gap: 0.18rem;
    min-width: 0;
    flex: 1 1 auto;
  }

  .settings-field__label-row {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    min-width: 0;
  }

  .settings-field__copy strong {
    color: var(--text-primary);
    font-size: 0.8rem;
  }

  .settings-field__copy p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.72rem;
    line-height: 1.48;
  }

  .settings-field__badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.5rem;
    height: 1.5rem;
    padding: 0 0.36rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--warning) 14%, var(--surface-panel));
    color: color-mix(in srgb, var(--warning) 86%, var(--text-primary));
    font-size: 0.62rem;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .settings-field__action {
    flex: 0 0 auto;
  }

  .settings-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    min-height: 2.1rem;
    padding: 0.28rem 0.4rem 0.28rem 0.32rem;
    border-radius: 999px;
    background: var(--surface-panel);
    box-shadow: none;
  }

  .settings-toggle.is-active {
    border-color: color-mix(in srgb, var(--accent) 26%, var(--border-strong));
    background: color-mix(in srgb, var(--accent) 9%, var(--surface-panel));
  }

  .settings-toggle__track {
    position: relative;
    width: 2.4rem;
    height: 1.5rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--accent) 16%, var(--surface-raised));
    box-shadow: inset 0 0 0 1px var(--border-soft);
  }

  .settings-toggle__thumb {
    position: absolute;
    top: 0.14rem;
    left: 0.16rem;
    width: 1.22rem;
    height: 1.22rem;
    border-radius: 999px;
    background: var(--surface-raised);
    box-shadow: 0 4px 12px color-mix(in srgb, black 12%, transparent);
    transition: transform 160ms ease;
  }

  .settings-toggle.is-active .settings-toggle__thumb {
    transform: translateX(0.86rem);
  }

  .settings-toggle small,
  .settings-checkbox span {
    color: var(--text-secondary);
    font-size: 0.7rem;
    font-weight: 700;
  }

  .settings-checkbox {
    display: inline-flex;
    align-items: center;
    gap: 0.44rem;
    min-height: 2rem;
  }

  .settings-choice {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.22rem;
    border-radius: 999px;
    border: 1px solid var(--border-soft);
    background: var(--surface-panel);
  }

  .settings-choice__option {
    display: inline-flex;
    align-items: center;
    gap: 0.34rem;
    min-height: 1.92rem;
    padding: 0.3rem 0.62rem;
    border-radius: 999px;
    box-shadow: none;
    background: transparent;
    border-color: transparent;
    color: var(--text-secondary);
  }

  .settings-choice__option.is-active {
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-raised));
    border-color: color-mix(in srgb, var(--accent) 22%, var(--border-strong));
    color: color-mix(in srgb, var(--accent) 84%, var(--text-primary));
  }

  .settings-choice__glyph {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 1.2rem;
    height: 1.2rem;
    border-radius: 999px;
    background: var(--surface-raised);
    font-size: 0.6rem;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .settings-input {
    min-width: 220px;
  }

  @media (max-width: 760px) {
    .settings-field {
      align-items: stretch;
      flex-direction: column;
    }

    .settings-field__action {
      width: 100%;
    }

    .settings-choice {
      width: 100%;
      justify-content: space-between;
      flex-wrap: wrap;
      border-radius: 22px;
    }
  }
</style>
