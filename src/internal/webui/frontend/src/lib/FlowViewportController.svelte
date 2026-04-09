<svelte:options runes={true} />

<script lang="ts">
  import { tick } from 'svelte';
  import { useSvelteFlow } from '@xyflow/svelte';

  let { request, padding, maxZoom }: { request: number; padding: number; maxZoom: number } = $props();

  const flow = useSvelteFlow();
  let lastAppliedRequest = 0;

  $effect(() => {
    const currentRequest = request;
    if (currentRequest === 0 || currentRequest === lastAppliedRequest) {
      return;
    }

    void refocus(currentRequest);
  });

  async function refocus(currentRequest: number): Promise<void> {
    await tick();
    await flow.fitView({
      padding,
      duration: 260,
      maxZoom,
    });
    lastAppliedRequest = currentRequest;
  }
</script>
