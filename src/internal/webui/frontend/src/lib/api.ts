import type { BackendGraph, VisualizationExportDocument } from './types';

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
  mode?: VisualizationExportDocument['meta']['mode'];
};

export async function loadGraphFromApi(fetcher: typeof fetch = fetch): Promise<VisualizationExportDocument> {
  return await loadVisualizationExport('/api/export', fetcher);
}

export async function loadVisualizationExport(url: string, fetcher: typeof fetch = fetch): Promise<VisualizationExportDocument> {
  const response = await fetcher(url, { cache: 'no-store' });
  if (!response.ok) {
    throw new ExportLoadError(`export request failed: ${response.status}`, response.status);
  }
  return parseVisualizationExport(await response.json(), {
    sourceLabel: sourceLabelForURL(url),
    mode: 'offline',
  });
}

export function parseVisualizationExport(
  input: unknown,
  fallback: CanonicalFallbackOptions,
): VisualizationExportDocument {
  if (isVisualizationExport(input)) {
    return input;
  }
  if (isGraphPayload(input)) {
    return wrapGraphAsVisualization(input, fallback);
  }
  throw new ExportLoadError('unsupported JSON format: expected a visualization export or graph JSON');
}

export function isNotFoundError(error: unknown): boolean {
  return error instanceof ExportLoadError && error.status === 404;
}

function isVisualizationExport(input: unknown): input is VisualizationExportDocument {
  if (!input || typeof input !== 'object') {
    return false;
  }
  const value = input as Partial<VisualizationExportDocument>;
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

function wrapGraphAsVisualization(graph: BackendGraph, fallback: CanonicalFallbackOptions): VisualizationExportDocument {
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
      tags: [],
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
