type ApiMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

type ApiContent =
  | Blob
  | File
  | Pick<ReadableStreamDefaultReader<any>, "read">
  | "";

interface ApiOpts {
  method?: ApiMethod;
  headers?: object;
  body?: any;
  signal?: AbortSignal;
}

interface TusSettings {
  retryCount: number;
  chunkSize: number;
}

type ChecksumAlg = "md5" | "sha1" | "sha256" | "sha512";

interface Share {
  hash: string;
  path: string;
  expire?: any;
  userID?: number;
  token?: string;
  username?: string;
}

type GrantRole = "viewer" | "editor";

interface Grant {
  id: number;
  path: string;
  ownerID: number;
  granteeID: number;
  role: GrantRole;
  createdBy: number;
  createdAt: number;
  expire: number;
  ownerUsername?: string;
  granteeUsername?: string;
}

interface GrantUser {
  id: number;
  username: string;
  displayName: string;
}

interface SearchParams {
  [key: string]: string;
}
