import type { BackendGraph, CanonicalExportDocument } from './types';

export class ExportLoadError extends Error {
  status?: number;

  constructor(message: string, status?: number) {
    super(message);
    this.name = 'ExportLoadError';
    this.status = status;
  }
}

type CanonicalFallbackOptions = {
  sourceLabel: string;
  mode?: CanonicalExportDocument['meta']['mode'];
};

export async function loadGraphFromApi(fetcher: typeof fetch = fetch): Promise<CanonicalExportDocument> {
  return await loadCanonicalExport('/api/export', fetcher);
}

export async function loadCanonicalExport(url: string, fetcher: typeof fetch = fetch): Promise<CanonicalExportDocument> {
  const response = await fetcher(url, { cache: 'no-store' });
  if (!response.ok) {
    throw new ExportLoadError(`export request failed: ${response.status}`, response.status);
  }
  return parseCanonicalExport(await response.json(), {
    sourceLabel: sourceLabelForURL(url),
    mode: 'offline',
  });
}

export function parseCanonicalExport(
  input: unknown,
  fallback: CanonicalFallbackOptions,
): CanonicalExportDocument {
  if (isCanonicalExport(input)) {
    return input;
  }
  if (isGraphPayload(input)) {
    return wrapGraphAsCanonical(input, fallback);
  }
  throw new ExportLoadError('unsupported JSON format: expected a canonical export or graph JSON');
}

export function isNotFoundError(error: unknown): boolean {
  return error instanceof ExportLoadError && error.status === 404;
}

function isCanonicalExport(input: unknown): input is CanonicalExportDocument {
  if (!input || typeof input !== 'object') {
    return false;
  }
  const value = input as Partial<CanonicalExportDocument>;
  return typeof value.schemaVersion === 'number'
    && typeof value.generatedAt === 'string'
    && typeof value.toolVersion === 'string'
    && value.graph !== undefined
    && value.source !== undefined
    && value.catalog !== undefined
    && value.validation !== undefined
    && value.meta !== undefined;
}

function isGraphPayload(input: unknown): input is BackendGraph {
  if (!input || typeof input !== 'object') {
    return false;
  }
  const value = input as Partial<BackendGraph>;
  return typeof value.schemaVersion === 'number'
    && typeof value.metadata === 'object'
    && Array.isArray(value.nodes)
    && Array.isArray(value.edges);
}

function wrapGraphAsCanonical(graph: BackendGraph, fallback: CanonicalFallbackOptions): CanonicalExportDocument {
  return {
    schemaVersion: 1,
    generatedAt: graph.metadata?.generatedAt ?? new Date().toISOString(),
    toolVersion: graph.metadata?.scannerVersion ?? 'unknown',
    source: {
      projectRoot: graph.metadata?.sourceRoot ?? '.',
      configPath: '',
      scopes: [],
    },
    catalog: {
      teams: [],
      domains: [],
    },
    validation: {
      diagnostics: [],
      summary: {
        errors: 0,
        warnings: 0,
        nodes: graph.nodes?.length ?? 0,
        edges: graph.edges?.length ?? 0,
      },
    },
    graph,
    ui: {},
    meta: {
      sourceLabel: fallback.sourceLabel,
      mode: fallback.mode ?? 'offline',
    },
  };
}

function sourceLabelForURL(url: string): string {
  if (url === '/api/export') {
    return 'live api';
  }
  if (url === './data.json') {
    return 'static build';
  }
  return `url: ${url}`;
}
