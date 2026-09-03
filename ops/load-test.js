import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 50,
    duration: '30s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/api/v1';
const SHOP_ID = __ENV.SHOP_ID || '11111111-aaaa-bbbb-cccc-000000000001';

export default function () {
    const requestHeaders = {
        'Content-Type': 'application/json',
        'X-Shop-ID': SHOP_ID,
    };

    const catalogResponse = http.get(`${BASE_URL}/catalog`, { headers: requestHeaders });
    check(catalogResponse, {
        'catalog status is 200': (res) => res.status === 200,
    });

    sleep(1);

    const checkoutPayload = JSON.stringify({
        customer_email: 'loadtest@example.com',
        payment_method: 'cod',
    });

    const checkoutResponse = http.post(`${BASE_URL}/checkout/intent`, checkoutPayload, {
        headers: requestHeaders,
    });
    check(checkoutResponse, {
        'checkout intent handled': (res) => res.status === 200 || res.status === 201 || res.status === 400,
    });

    sleep(2);
}
