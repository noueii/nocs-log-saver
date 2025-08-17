const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:9090';

export interface Server {
  id: string;
  name: string;
  ip_address: string;
  api_key: string;
  description: string;
  is_active: boolean;
  last_seen: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface Log {
  id: string;
  server_id: string;
  content: string;
  created_at: string;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.statusText}`);
    }

    return response.json();
  }

  // Server Management
  async getServers(): Promise<Server[]> {
    const token = localStorage.getItem('access_token');
    return this.request('/api/admin/servers', {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
  }

  async getServer(id: string): Promise<Server> {
    const token = localStorage.getItem('access_token');
    return this.request(`/api/admin/servers/${id}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
  }

  async createServer(server: { name: string; description: string }): Promise<Server> {
    const token = localStorage.getItem('access_token');
    return this.request('/api/admin/servers', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(server),
    });
  }

  async updateServer(id: string, server: { name: string; description: string; is_active: boolean }): Promise<Server> {
    const token = localStorage.getItem('access_token');
    return this.request(`/api/admin/servers/${id}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(server),
    });
  }

  async deleteServer(id: string): Promise<void> {
    const token = localStorage.getItem('access_token');
    await this.request(`/api/admin/servers/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
  }

  async regenerateApiKey(id: string): Promise<{ api_key: string; message: string }> {
    const token = localStorage.getItem('access_token');
    return this.request(`/api/admin/servers/${id}/regenerate-key`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
  }


  // Logs
  async getLogs(serverId?: string, type: 'raw' | 'parsed' = 'raw'): Promise<Log[]> {
    const params = new URLSearchParams();
    if (serverId) params.append('server_id', serverId);
    params.append('type', type);
    
    return this.request(`/api/logs?${params.toString()}`);
  }

  // Stats
  async getStats(): Promise<any> {
    return this.request('/api/stats');
  }
}

export const api = new ApiClient();