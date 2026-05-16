interface Env {
  ASSETS: Fetcher;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    let { pathname } = url;

    // Trailing slash → serve index.html
    if (pathname.endsWith("/")) {
      pathname += "index.html";
    }

    // Clean URLs: /foo → /foo.html if no extension
    if (!pathname.includes(".")) {
      pathname += ".html";
    }

    return env.ASSETS.fetch(new Request(pathname, request));
  },
};