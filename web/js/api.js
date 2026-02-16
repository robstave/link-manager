const API_BASE = '/api/v1';

class ApiClient {
    constructor() {
        this.token = localStorage.getItem('token');
    }

    async request(path, options = {}) {
        const url = `${API_BASE}${path}`;
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const resp = await fetch(url, { ...options, headers });

        if (resp.status === 401) {
            this.logout();
            return null;
        }

        if (!resp.ok) {
            const error = await resp.text();
            throw new Error(error || 'Request failed');
        }

        if (resp.status === 204) return null;
        return resp.json();
    }

    async login(username, password) {
        const data = await this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });

        if (data && data.token) {
            this.token = data.token;
            localStorage.setItem('token', data.token);
            return data;
        }
        return null;
    }

    logout() {
        this.token = null;
        localStorage.removeItem('token');
        window.location.reload();
    }

    isAuthenticated() {
        return !!this.token;
    }

    // Projects
    getProjects() {
        return this.request('/projects');
    }

    createProject(name, description) {
        return this.request('/projects', {
            method: 'POST',
            body: JSON.stringify({ name, description }),
        });
    }

    // Categories
    getCategories(projectId) {
        return this.request(`/projects/${projectId}/categories`);
    }

    createCategory(projectId, name) {
        return this.request(`/projects/${projectId}/categories`, {
            method: 'POST',
            body: JSON.stringify({ name }),
        });
    }

    // Links
    getLinks(params = {}) {
        const query = new URLSearchParams(params).toString();
        return this.request(`/links?${query}`);
    }

    createLink(linkData) {
        return this.request('/links', {
            method: 'POST',
            body: JSON.stringify(linkData),
        });
    }

    recordClick(linkId) {
        return this.request(`/links/${linkId}/click`, { method: 'POST' });
    }

    updateStars(linkId, stars) {
        return this.request(`/links/${linkId}/stars`, {
            method: 'PATCH',
            body: JSON.stringify({ stars }),
        });
    }

    // Tags
    getTags() {
        return this.request('/tags');
    }

    getMe() {
        return this.request('/auth/me');
    }
}

const api = new ApiClient();
