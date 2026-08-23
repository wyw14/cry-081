import type { APIError, ManuscriptPage, PublicationPage } from '../types/editorial'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

export class RequestError extends Error {
  constructor(public readonly detail: APIError, public readonly status: number) {
    super(detail.message)
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${baseURL}${path}`, {
    ...options,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      'X-Actor-ID': sessionStorage.getItem('actor_id') ?? 'demo-editor',
      ...(options.headers ?? {})
    }
  })
  if (!response.ok) {
    const detail = (await response.json()) as APIError
    throw new RequestError(detail, response.status)
  }
  return response.json() as Promise<T>
}

export const editorialAPI = {
  listManuscripts(params: URLSearchParams): Promise<ManuscriptPage> {
    return request(`/manuscripts?${params.toString()}`)
  },
  searchPublications(params: URLSearchParams): Promise<PublicationPage> {
    return request(`/publications?${params.toString()}`)
  }
}
