import http from 'k6/http';
import { check, sleep } from 'k6';
import { jwtHS256, randomUser } from './lib.js';

const BASE = __ENV.BASE_URL || 'http://gateway:8088';
const SCENARIO = __ENV.SCENARIO || 'smoke';
const TARGET_QPS = Number(__ENV.TARGET_QPS || 200);
const DURATION = __ENV.TEST_DURATION || '30s';
const VUS_MAX = Number(__ENV.VUS_MAX || 200);
const HOT_FEED_LIMIT = Number(__ENV.HOT_FEED_LIMIT || 50);
const SECRET = __ENV.JWT_SECRET || 'dev-secret';
const POST_IDS = (__ENV.POST_IDS || '').split(',').filter(Boolean);

export const options = {
  scenarios: {
    main: {
      executor: 'constant-arrival-rate',
      rate: TARGET_QPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: Math.min(50, VUS_MAX),
      maxVUs: VUS_MAX,
      exec: 'run',
    },
  },
  thresholds: {
    http_req_failed: [`rate<${__ENV.FAIL_ERROR_RATE || '0.01'}`],
    http_req_duration: [`p(95)<${__ENV.FAIL_P95_MS || '500'}`, `p(99)<${__ENV.FAIL_P99_MS || '1000'}`],
  },
};

function pickPostId() {
  if (POST_IDS.length === 0) return null;
  return POST_IDS[Math.floor(Math.random() * POST_IDS.length)];
}

function createPost() {
  const user = randomUser();
  const token = jwtHS256(SECRET, user);
  const payload = JSON.stringify({ authorId: user, content: `k6-${Date.now()}-${Math.random()}`, tags: ['bench', 'k6'] });
  const r = http.post(`${BASE}/api/v1/posts`, payload, { headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` } });
  check(r, { 'create post 201': (res) => res.status === 201 });
}

function getPost() {
  const id = pickPostId();
  if (!id) return;
  const r = http.get(`${BASE}/api/v1/posts/${id}`);
  check(r, { 'get post 200': (res) => res.status === 200 || res.status === 404 });
}

function getHot() {
  const r = http.get(`${BASE}/api/v1/feed/hot?limit=${HOT_FEED_LIMIT}`);
  check(r, { 'hot feed 200': (res) => res.status === 200 });
}

function likePost() {
  const id = pickPostId();
  if (!id) return;
  const user = randomUser();
  const token = jwtHS256(SECRET, user);
  const r = http.post(`${BASE}/api/v1/posts/${id}/like`, null, { headers: { Authorization: `Bearer ${token}` } });
  check(r, { 'like 200': (res) => res.status === 200 });
}

export function run() {
  const dice = Math.random();
  if (SCENARIO === 'smoke') {
    getHot(); getPost(); createPost(); sleep(0.05); return;
  }
  if (SCENARIO === 'read-heavy') {
    if (dice < 0.75) getHot(); else getPost();
    return;
  }
  // mixed: 75% hot, 15% post, 5% create, 5% like
  if (dice < 0.75) getHot();
  else if (dice < 0.90) getPost();
  else if (dice < 0.95) createPost();
  else likePost();
}
