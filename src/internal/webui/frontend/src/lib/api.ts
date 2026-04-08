import type { ExplorerPayload } from './types';

export async function loadGraphFromApi(fetcher: typeof fetch = fetch): Promise<ExplorerPayload> {
  const explorerResponse = await fetcher('/api/explorer');

  if (!explorerResponse.ok) {
    throw new Error(`explorer request failed: ${explorerResponse.status}`);
  }
  return await explorerResponse.json() as ExplorerPayload;
}
