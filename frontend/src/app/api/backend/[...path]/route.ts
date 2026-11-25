type RouteContext = {
  params: {
    path: string[];
  };
};

export const dynamic = 'force-dynamic';

async function proxy(request: Request, { params }: RouteContext): Promise<Response> {
  const configuredURL = process.env.BACKEND_API_URL || process.env.NEXT_PUBLIC_API_URL;
  if (!configuredURL) {
    return new Response('Backend API URL is not configured', { status: 503 });
  }

  let backendURL: URL;
  try {
    backendURL = new URL(configuredURL.includes('://') ? configuredURL : `https://${configuredURL}`);
  } catch {
    return new Response('Backend API URL is invalid', { status: 503 });
  }

  const path = params.path.map((part) => encodeURIComponent(part)).join('/');
  backendURL.pathname = `${backendURL.pathname.replace(/\/+$/, '')}/${path}`;
  backendURL.search = new URL(request.url).search;

  const headers = new Headers();
  for (const header of ['accept', 'content-type']) {
    const value = request.headers.get(header);
    if (value) headers.set(header, value);
  }
  headers.set('User-Agent', 'DineWare-Operator-Portal/1.0');
  headers.set('Accept', 'application/json');

  try {
    const response = await fetch(backendURL, {
      method: request.method,
      headers,
      body: request.method === 'GET' || request.method === 'HEAD'
        ? undefined
        : await request.arrayBuffer(),
      cache: 'no-store',
    });
    const responseHeaders = new Headers();
    const contentType = response.headers.get('content-type');
    if (contentType) responseHeaders.set('content-type', contentType);
    responseHeaders.set('cache-control', 'no-store');
    return new Response(response.body, {
      status: response.status,
      headers: responseHeaders,
    });
  } catch {
    return new Response('Backend API is unavailable', { status: 502 });
  }
}

export const GET = proxy;
export const POST = proxy;
export const PATCH = proxy;
export const PUT = proxy;
export const DELETE = proxy;
