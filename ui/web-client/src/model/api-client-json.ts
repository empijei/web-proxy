export class HTTPError extends Error {
  constructor(
    public readonly status: number,
    public readonly message: string,
  ) {
    super(message);
    this.name = 'HTTPError';
  }
}

export type ProcedureJSON<ResponseT, RequestT> = (
  req: RequestT,
  sig?: AbortSignal,
) => Promise<ResponseT>;

export function RemoteJSONEndpoint<ResponseT, RequestT>(
  method: string,
  origin: string,
  path: string,
  authorization?: string,
): ProcedureJSON<ResponseT, RequestT> {
  const stateChanging = !(
    method === 'GET' ||
    method === 'HEAD' ||
    method === 'OPTIONS'
  );

  return async function (req: RequestT, sig?: AbortSignal): Promise<ResponseT> {
    const params = typeof req !== 'undefined' ? JSON.stringify(req) : '';
    let response: Response;
    const url = new URL(path, origin || window.location.origin);

    if (stateChanging) {
      response = await fetch(url, {
        method: method,
        body: params,
        headers: { 'Content-Type': 'application/json' },
        signal: sig,
      });
    } else {
      url.searchParams.set('srpc', JSON.stringify(req));
      response = await fetch(url, {
        method: method,
        signal: sig,
      });
    }

    if (response.ok) {
      return response.json() as Promise<ResponseT>;
    }

    const body = await response.text();
    throw new HTTPError(
      response.status,
      body.trim() || `HTTP ${response.status} error`,
    );
  };
}
