const baseURL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';
export class RequestError extends Error {
    detail;
    status;
    constructor(detail, status) {
        super(detail.message);
        this.detail = detail;
        this.status = status;
    }
}
async function request(path, options = {}) {
    const response = await fetch(`${baseURL}${path}`, {
        ...options,
        headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
            'X-Actor-ID': sessionStorage.getItem('actor_id') ?? 'demo-editor',
            ...(options.headers ?? {})
        }
    });
    if (!response.ok) {
        const detail = (await response.json());
        throw new RequestError(detail, response.status);
    }
    return response.json();
}
export const editorialAPI = {
    listManuscripts(params) {
        return request(`/manuscripts?${params.toString()}`);
    },
    searchPublications(params) {
        return request(`/publications?${params.toString()}`);
    }
};
