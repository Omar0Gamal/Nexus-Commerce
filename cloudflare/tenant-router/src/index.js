export default {
	async fetch(request, env, ctx) {
		const url = new URL(request.url);
		const hostname = url.hostname;

		// 1. Resolve Tenant
		// Check the KV store for the hostname (e.g. shop.com -> tenant_id)
		let tenantId = await env.TENANT_KV.get(hostname);
		
		// Fallback for subdomains like shopname.egyptcommerce.com
		if (!tenantId && hostname.endsWith('.egyptcommerce.com')) {
			const subdomain = hostname.split('.')[0];
			tenantId = await env.TENANT_KV.get(`subdomain:${subdomain}`);
		}

		// 2. Clone the request so we can modify headers
		const modifiedRequest = new Request(request);

		// Inject the resolved Shop ID as a trusted header
		if (tenantId) {
			modifiedRequest.headers.set('X-Shop-ID', tenantId);
		}

		// 3. Smart Routing
		// If the request is for the API, route to the backend origin
		if (url.pathname.startsWith('/api/')) {
			const backendUrl = new URL(url);
			// Example: Your Go backend might run on api.egyptcommerce.com in prod
			backendUrl.hostname = env.BACKEND_ORIGIN; 
			return fetch(new Request(backendUrl, modifiedRequest));
		}

		// Otherwise, route to the Next.js frontend origin
		const frontendUrl = new URL(url);
		// Example: Your Next.js app might run on storefront.egyptcommerce.com
		frontendUrl.hostname = env.FRONTEND_ORIGIN;
		return fetch(new Request(frontendUrl, modifiedRequest));
	},
};
