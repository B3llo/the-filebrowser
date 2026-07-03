import { fetchURL, fetchJSON, StatusError, createURL } from "./utils";

export async function getAll() {
  return fetchJSON<IUser[]>(`/api/users`, {});
}

export async function get(id: number) {
  return fetchJSON<IUser>(`/api/users/${id}`, {});
}

export async function create(user: IUser, currentPassword: string) {
  const res = await fetchURL(`/api/users`, {
    method: "POST",
    body: JSON.stringify({
      what: "user",
      which: [],
      current_password: currentPassword,
      data: user,
    }),
  });

  if (res.status === 201) {
    return res.headers.get("Location");
  }

  throw new StatusError(await res.text(), res.status);
}

export async function update(
  user: Partial<IUser>,
  which = ["all"],
  currentPassword: string | null = null
) {
  await fetchURL(`/api/users/${user.id}`, {
    method: "PUT",
    body: JSON.stringify({
      what: "user",
      which: which,
      ...(currentPassword != null ? { current_password: currentPassword } : {}),
      data: user,
    }),
  });
}

export async function remove(
  id: number,
  currentPassword: string | null = null
) {
  await fetchURL(`/api/users/${id}`, {
    method: "DELETE",
    body: JSON.stringify({
      ...(currentPassword != null ? { current_password: currentPassword } : {}),
    }),
  });
}

export async function uploadAvatar(file: File) {
  await fetchURL(`/api/avatar`, {
    method: "POST",
    body: file,
  });
}

export async function removeAvatar() {
  await fetchURL(`/api/avatar`, {
    method: "DELETE",
  });
}

// bust lets callers force a fresh fetch right after an upload, since the
// server always serves the same URL for a given user id (Cache-Control is
// short-lived but a same-second reload could still hit the browser cache).
export function avatarURL(id: number, bust?: number): string {
  return createURL(`api/avatar/${id}`, bust ? { v: bust } : {});
}
