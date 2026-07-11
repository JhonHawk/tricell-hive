declare const process: {
  env: {
    DOCS_BASIC_AUTH_PASSWORD?: string;
    DOCS_BASIC_AUTH_USER?: string;
  };
};

const unauthorizedResponse = new Response("Authentication required", {
  status: 401,
  headers: {
    "WWW-Authenticate": 'Basic realm="Docs", charset="UTF-8"',
  },
});

const missingConfigResponse = new Response("Authentication is not configured", {
  status: 503,
});

function constantTimeEqual(left: string, right: string): boolean {
  const maxLength = Math.max(left.length, right.length);
  let mismatch = left.length === right.length ? 0 : 1;

  for (let index = 0; index < maxLength; index += 1) {
    mismatch |= (left.charCodeAt(index) || 0) ^ (right.charCodeAt(index) || 0);
  }

  return mismatch === 0;
}

function readBasicCredentials(request: Request): { password: string; user: string } | null {
  const authorization = request.headers.get("authorization");

  if (!authorization?.startsWith("Basic ")) {
    return null;
  }

  try {
    const decoded = atob(authorization.slice("Basic ".length));
    const separatorIndex = decoded.indexOf(":");

    if (separatorIndex < 0) {
      return null;
    }

    return {
      user: decoded.slice(0, separatorIndex),
      password: decoded.slice(separatorIndex + 1),
    };
  } catch {
    return null;
  }
}

export default function middleware(request: Request): Response | undefined {
  const expectedUser = process.env.DOCS_BASIC_AUTH_USER;
  const expectedPassword = process.env.DOCS_BASIC_AUTH_PASSWORD;

  if (!expectedUser || !expectedPassword) {
    return missingConfigResponse;
  }

  const credentials = readBasicCredentials(request);

  if (
    !credentials ||
    !constantTimeEqual(credentials.user, expectedUser) ||
    !constantTimeEqual(credentials.password, expectedPassword)
  ) {
    return unauthorizedResponse;
  }

  return undefined;
}
