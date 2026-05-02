import http from 'k6/http';
import { check } from 'k6';

const baseURL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const vus = Number(__ENV.VUS || 1);
const iterations = Number(__ENV.ITERATIONS || 20);
const maxDuration = __ENV.MAX_DURATION || '30s';

export const options = {
  scenarios: {
    root_smoke: {
      executor: 'shared-iterations',
      vus,
      iterations,
      maxDuration,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  const res = http.get(`${baseURL}/`);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'success is true': (r) => r.json('success') === true,
    'code is 200': (r) => r.json('code') === 200,
    'message matches': (r) => r.json('message') === 'Welcome to Fiber Starter API',
    'api version link present': (r) => r.json('data.api') === '/api/v1',
    'docs link present': (r) => r.json('data.docs') === '/docs',
  });
}
