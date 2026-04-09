<script lang="ts">
  let {
    title,
    description,
    children,
  }: {
    title: string;
    description: string;
    children?: () => unknown;
  } = $props();
</script>

<section class="settings-block">
  <header class="settings-block__head">
    <div>
      <strong>{title}</strong>
      <p>{description}</p>
    </div>
  </header>

  <div class="settings-block__body">
    {@render children?.()}
  </div>
</section>

<style>
  .settings-block {
    display: grid;
    gap: 0.8rem;
    padding: 0.9rem;
    border-radius: 24px;
    border: 1px solid var(--border-soft);
    background: var(--surface-panel-soft);
  }

  .settings-block__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.9rem;
  }

  .settings-block__head strong {
    display: block;
    margin-bottom: 0.18rem;
    color: var(--text-primary);
    font-size: 0.92rem;
  }

  .settings-block__head p {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.74rem;
    line-height: 1.5;
  }

  .settings-block__body {
    display: grid;
    gap: 0.6rem;
  }
</style>
