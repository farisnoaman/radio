import simpleRestProvider from 'ra-data-simple-rest';
import {
  fetchUtils,
  DataProvider,
  Identifier,
  RaRecord,
  CreateResult,
  CreateParams,
} from 'react-admin';
import { API_BASE, extractData, extractTotal, httpClient } from '../utils/apiClient';

// Get current tenant ID from localStorage (kept for future use)
const getTenantID = (): string | null => {
  const userStr = localStorage.getItem('user');
  if (userStr) {
    try {
      const user = JSON.parse(userStr);
      return user.tenant_id ? String(user.tenant_id) : null;
    } catch {
      return null;
    }
  }
  return null;
};

// Mark as used to avoid unused variable warning
void getTenantID;

const getIdentifier = (payload: unknown): string | number | undefined => {
  if (typeof payload === 'object' && payload !== null && 'id' in payload) {
    const value = (payload as { id?: string | number }).id;
    if (typeof value === 'string' || typeof value === 'number') {
      return value;
    }
  }
  return undefined;
};

const resourcePathMap: Record<string, string> = {
  'radius/users': 'users',
  'radius/online': 'sessions',
  'radius/accounting': 'accounting',
  'radius/profiles': 'radius-profiles',
  'system/config/schemas': 'system/config/schemas',
  'campaigns': 'campaigns',
  'cpes': 'cpes',
  // Phase 4: Monitoring
  'monitoring/devices': 'monitoring/devices',
  'monitoring/metrics': 'monitoring/metrics',
  // Phase 5A: Billing
  'billing/invoices': 'billing/invoices',
  'billing/plans': 'admin/billing/plans',
  // Phase 5B: Backups
  'provider/backup': 'provider/backup',
  // Platform Management
  'providers/registrations': 'providers/registrations',
  'platform/settings': 'admin/platform/settings',
  'admin/providers': 'admin/providers',
  'admin/monitoring/provider': 'admin/monitoring/provider',
  'quotas': 'admin/providers',
  // Network
  'network/locations': 'locations',
  'network/devices': 'network/devices',
};

const resolveResource = (resource: string) =>
  resourcePathMap[resource] ?? resource;

const buildApiUrl = (resource: string, suffix = '') =>
  `${API_BASE}/${resolveResource(resource)}${suffix}`;

const baseDataProvider = simpleRestProvider(API_BASE, httpClient);

// 自定义 dataProvider 以适配后端 API 格式
export const dataProvider: DataProvider = {
  ...baseDataProvider,

  getList: async (resource, params) => {
    const { page = 1, perPage = 10 } = params.pagination || {};
    const { field = 'id', order = 'ASC' } = params.sort || {};
    const query = {
      sort: field,
      order,
      page,
      perPage,
      ...params.filter,
    };

    const url = `${buildApiUrl(resource)}?${fetchUtils.queryParameters(query)}`;
    const { json } = await httpClient(url);

    return {
      data: extractData(json),
      total: extractTotal(json),
    };
  },

  getOne: async (resource, params) => {
    const url = buildApiUrl(resource, `/${params.id}`);
    const { json } = await httpClient(url);
    return { data: extractData(json) };
  },

  getMany: async (resource, params) => {
    const query = {
      filter: JSON.stringify({ id: params.ids }),
    };
    const url = `${buildApiUrl(resource)}?${fetchUtils.queryParameters(query)}`;
    const { json } = await httpClient(url);
    return { data: extractData(json) };
  },

  getManyReference: async (resource, params) => {
    const { page = 1, perPage = 10 } = params.pagination || {};
    const { field = 'id', order = 'ASC' } = params.sort || {};
    const query = {
      sort: field,
      order: order,
      page: page,
      perPage,
      ...params.filter,
      [params.target]: params.id,
    };
    const url = `${buildApiUrl(resource)}?${fetchUtils.queryParameters(query)}`;
    const { json } = await httpClient(url);
    return {
      data: extractData(json),
      total: extractTotal(json),
    };
  },

  create: async function create<RecordType extends Omit<RaRecord, 'id'> = Omit<RaRecord, 'id'>, ResultRecordType extends RaRecord = RecordType & { id: Identifier }>(
    resource: string,
    params: CreateParams<RecordType>
  ): Promise<CreateResult<ResultRecordType>> {
    const url = buildApiUrl(resource);
    const { json } = await httpClient(url, {
      method: 'POST',
      body: JSON.stringify(params.data),
    });
    const data = extractData<Record<string, unknown> | undefined>(json);
    const resolvedId =
      getIdentifier(data) ?? getIdentifier(json) ?? getIdentifier(params.data);

    if (resolvedId === undefined) {
      throw new Error('Create response missing identifier');
    }

    const basePayload =
      typeof params.data === 'object' && params.data !== null ? params.data : ({} as RecordType);

    const mergedPayload = {
      ...basePayload,
      ...(data ?? {}),
      id: resolvedId,
    } satisfies RaRecord;

    return { data: mergedPayload as unknown as ResultRecordType };
  },

  update: async (resource, params) => {
    const url = buildApiUrl(resource, `/${params.id}`);
    const { json } = await httpClient(url, {
      method: 'PUT',
      body: JSON.stringify(params.data),
    });
    return { data: extractData(json) };
  },

  updateMany: async (resource, params) => {
    const responses = await Promise.all(
      params.ids.map(id =>
        httpClient(buildApiUrl(resource, `/${id}`), {
          method: 'PUT',
          body: JSON.stringify(params.data),
        })
      )
    );
    return {
      data: responses.map(({ json }, index) => {
        const data = extractData<Record<string, unknown> | undefined>(json);
        return getIdentifier(data) ?? getIdentifier(json) ?? params.ids[index];
      }),
    };
  },

  delete: async (resource, params) => {
    const url = buildApiUrl(resource, `/${params.id}`);
    const { json } = await httpClient(url, {
      method: 'DELETE',
    });
    return { data: extractData(json) };
  },

  deleteMany: async (resource, params) => {
    const responses = await Promise.all(
      params.ids.map(id =>
        httpClient(buildApiUrl(resource, `/${id}`), {
          method: 'DELETE',
        })
      )
    );
    return {
      data: responses.map(({ json }, index) => {
        const data = extractData<Record<string, unknown> | undefined>(json);
        return getIdentifier(data) ?? getIdentifier(json) ?? params.ids[index];
      }),
    };
  },
};
