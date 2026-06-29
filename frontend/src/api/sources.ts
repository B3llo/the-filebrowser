import { fetchJSON, fetchURL } from "./utils";

export const sources = {
  async list(): Promise<ISource[]> {
    return fetchJSON<ISource[]>("/api/sources");
  },

  async get(id: number): Promise<ISource> {
    return fetchJSON<ISource>(`/api/sources/${id}`);
  },

  async create(data: { name: string; path: string }): Promise<Response> {
    return fetchURL("/api/sources", {
      method: "POST",
      body: JSON.stringify({ what: "source", data }),
      headers: { "Content-Type": "application/json" },
    });
  },

  async update(
    id: number,
    fields: string[],
    data: Partial<ISource>
  ): Promise<Response> {
    return fetchURL(`/api/sources/${id}`, {
      method: "PUT",
      body: JSON.stringify({ what: "source", which: fields, data }),
      headers: { "Content-Type": "application/json" },
    });
  },

  async remove(id: number): Promise<Response> {
    return fetchURL(`/api/sources/${id}`, { method: "DELETE" });
  },
};
