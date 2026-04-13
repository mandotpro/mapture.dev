<script lang="ts">
  let {
    label,
    value = '',
    className = '',
    children,
  }: {
    label: string;
    value?: string;
    className?: string;
    children?: () => unknown;
  } = $props();
</script>

<div class={['ui-property-row', className].filter(Boolean).join(' ')}>
  <dt>{label}</dt>
  <dd>
    {#if children}
      {@render children()}
    {:else}
      {value}
    {/if}
  </dd>
</div>

<style>
  .ui-property-row {
    display: grid;
    grid-template-columns: minmax(84px, 108px) minmax(0, 1fr);
    gap: 0.8rem;
    padding-top: 0.62rem;
    border-top: 1px solid color-mix(in srgb, var(--border-soft) 88%, transparent);
  }

  .ui-property-row:first-child {
    padding-top: 0;
    border-top: 0;
  }

  .ui-property-row dt {
    color: var(--text-tertiary);
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  .ui-property-row dd {
    margin: 0;
    color: var(--text-primary);
    font-size: 0.79rem;
    line-height: 1.5;
    font-weight: 580;
    overflow-wrap: anywhere;
  }

  @media (max-width: 520px) {
    .ui-property-row {
      grid-template-columns: 1fr;
      gap: 0.18rem;
    }
  }
</style>
