<script lang="ts">
  import IconButton from './IconButton.svelte';

  let {
    open = false,
    title = '',
    description = '',
    width = 'min(720px, calc(100vw - 2rem))',
    onclose,
    children,
  }: {
    open?: boolean;
    title?: string;
    description?: string;
    width?: string;
    onclose?: () => void;
    children?: () => unknown;
  } = $props();

  function handleBackdropClick(event: MouseEvent): void {
    if (event.target !== event.currentTarget) {
      return;
    }
    onclose?.();
  }
</script>

{#if open}
  <div class="canvas-modal" role="presentation" onclick={handleBackdropClick}>
    <div
      class="canvas-modal__card"
      role="dialog"
      aria-modal="true"
      aria-label={title || 'Dialog'}
      style={`--canvas-modal-width:${width};`}
    >
      <header class="canvas-modal__header">
        <div class="canvas-modal__copy">
          {#if title}
            <strong>{title}</strong>
          {/if}
          {#if description}
            <p>{description}</p>
          {/if}
        </div>
        <IconButton className="canvas-modal__close" onclick={onclose} ariaLabel="Close dialog" subtle>
          <span class="canvas-modal__close-mark" aria-hidden="true">x</span>
        </IconButton>
      </header>

      <div class="canvas-modal__body">
        {@render children?.()}
      </div>
    </div>
  </div>
{/if}

<style>
  .canvas-modal {
    position: fixed;
    inset: 64px 0 0;
    z-index: 90;
    display: grid;
    place-items: center;
    padding: 1.2rem;
    background: rgba(10, 16, 24, 0.16);
    backdrop-filter: blur(8px);
  }

  .canvas-modal__card {
    width: var(--canvas-modal-width);
    max-height: min(78vh, 840px);
    display: grid;
    grid-template-rows: auto minmax(0, 1fr);
    overflow: hidden;
    border-radius: 30px;
    border: 1px solid var(--border-strong);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 11%, transparent), transparent 34%),
      var(--surface-overlay);
    box-shadow: var(--shadow-floating);
  }

  .canvas-modal__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    padding: 1.1rem 1.1rem 0.95rem;
    border-bottom: 1px solid var(--border-soft);
  }

  .canvas-modal__copy {
    display: grid;
    gap: 0.22rem;
    min-width: 0;
  }

  .canvas-modal__copy strong {
    font-family: "Iowan Old Style", "Palatino Linotype", serif;
    font-size: 1.18rem;
    color: var(--text-primary);
  }

  .canvas-modal__copy p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.8rem;
    line-height: 1.5;
  }

  .canvas-modal__close-mark {
    font-size: 0.75rem;
    font-weight: 800;
    text-transform: uppercase;
  }

  .canvas-modal__body {
    min-height: 0;
    overflow: auto;
    padding: 1rem 1.1rem 1.2rem;
  }

  @media (max-width: 920px) {
    .canvas-modal {
      inset: 78px 0 0;
      padding: 0.75rem;
    }

    .canvas-modal__card {
      width: min(100%, calc(100vw - 1rem));
      max-height: calc(100vh - 5.5rem);
      border-radius: 24px;
    }
  }
</style>
