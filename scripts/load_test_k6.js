import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    errors: ['rate<0.1'],              // Error rate should be less than 10%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = __ENV.API_URL || 'http://localhost:8081';

export default function () {
  // Test event ingestion
  const eventPayload = JSON.stringify({
    user_id: Math.floor(Math.random() * 1000) + 1,
    item_id: Math.floor(Math.random() * 100) + 1,
    event_type: ['VIEW', 'CLICK', 'CART', 'PURCHASE'][Math.floor(Math.random() * 4)],
    session_id: `session_${__VU}_${__ITER}`,
  });

  const eventRes = http.post(`${BASE_URL}/events`, eventPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(eventRes, {
    'event ingest status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);

  sleep(1);

  // Test recommendations API
  const userID = Math.floor(Math.random() * 1000) + 1;
  const recoRes = http.get(`${API_URL}/recommendations?user_id=${userID}&count=10`);

  check(recoRes, {
    'recommendations status is 200': (r) => r.status === 200,
    'recommendations has data': (r) => {
      const body = JSON.parse(r.body);
      return body.recommendations !== undefined;
    },
  }) || errorRate.add(1);

  sleep(1);

  // Test popular items API
  const popularRes = http.get(`${API_URL}/popular?count=20`);

  check(popularRes, {
    'popular status is 200': (r) => r.status === 200,
    'popular has recommendations': (r) => {
      const body = JSON.parse(r.body);
      return body.recommendations !== undefined;
    },
  }) || errorRate.add(1);

  sleep(2);
}
