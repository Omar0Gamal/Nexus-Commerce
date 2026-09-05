export default {
	async fetch(request, env) {
		const requestUrl = new URL(request.url);
		const hostname = requestUrl.hostname;

		let tenantId = await env.TENANT_KV.get(hostname);
		if (!tenantId && hostname.endsWith('.egyptcommerce.com')) {
			const subdomain = hostname.split('.')[0];
			tenantId = await env.TENANT_KV.get(`subdomain:${subdomain}`);
		}

		const upstreamRequest = new Request(request);
		if (tenantId) {
			upstreamRequest.headers.set('X-Shop-ID', tenantId);
		}
		// Shared secret verified by Caddy and Go middleware to reject direct-to-origin spoofing.
		upstreamRequest.headers.set('X-Gateway-Secret', env.GATEWAY_SECRET);

		const targetUrl = new URL(requestUrl);
		targetUrl.hostname = requestUrl.pathname.startsWith('/api/')
			? env.BACKEND_ORIGIN
			: env.FRONTEND_ORIGIN;

		return fetch(new Request(targetUrl, upstreamRequest));
	},
};
