import http from 'k6/http';
import { check } from 'k6';

const baseURL = (__ENV.BASE_URL || 'http://localhost:8080').replace(/\/$/, '');
const vus = Number(__ENV.VUS || 10);
const duration = __ENV.DURATION || '1m';
const targetDuration = Number(__ENV.TARGET_DURATION || 500);

export const options = {
  scenarios: {
    root_load: {
      executor: 'constant-vus',
      vus,
      duration,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: [`p(95)<${targetDuration}`],
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
