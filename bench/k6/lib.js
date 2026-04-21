import crypto from 'k6/crypto';
import encoding from 'k6/encoding';

function b64url(input) {
  return encoding.b64encode(input, 'rawstd').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function jwtHS256(secret, sub) {
  const now = Math.floor(Date.now() / 1000);
  const header = b64url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const payload = b64url(JSON.stringify({ sub, exp: now + 300, iat: now }));
  const data = `${header}.${payload}`;
  const sig = crypto.hmac('sha256', secret, data, 'base64rawurl');
  return `${data}.${sig}`;
}

export function randomUser() {
  return `u-${String(Math.floor(Math.random() * 200)).padStart(3, '0')}`;
}
