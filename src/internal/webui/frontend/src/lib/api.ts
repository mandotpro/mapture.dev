import type { AppPayload, CatalogPayload, ValidationPayload } from './types';

export async function loadGraphFromApi(fetcher: typeof fetch = fetch): Promise<AppPayload> {
  const [graphResponse, validationResponse, catalogResponse] = await Promise.all([
    fetcher('/api/graph'),
    fetcher('/api/validate'),
    fetcher('/api/catalog'),
  ]);

  if (!graphResponse.ok) {
    throw new Error(`graph request failed: ${graphResponse.status}`);
  }
  if (!validationResponse.ok) {
    throw new Error(`validate request failed: ${validationResponse.status}`);
  }
  if (!catalogResponse.ok) {
    throw new Error(`catalog request failed: ${catalogResponse.status}`);
  }

  return {
    graph: await graphResponse.json() as ValidationPayload,
    validation: await validationResponse.json() as ValidationPayload,
    catalog: await catalogResponse.json() as CatalogPayload,
  };
}

export async function loadGraphFromFile(file: File): Promise<ValidationPayload> {
  const text = await file.text();
  return JSON.parse(text) as ValidationPayload;
}
