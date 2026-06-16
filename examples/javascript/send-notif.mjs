/**
 * Contoh kirim notif via fetch (Node 18+)
 *
 * WHATSAR_URL=http://127.0.0.1:8080
 * WHATSAR_KEY=your-api-key
 * WHATSAR_SESSION_ID=uuid
 *
 * node send-notif.mjs 6281234567890 "Halo!"
 */

const baseUrl = process.env.WHATSAR_URL ?? 'http://127.0.0.1:8080';
const apiKey = process.env.WHATSAR_KEY;
const sessionId = process.env.WHATSAR_SESSION_ID;

if (!apiKey || !sessionId) {
  console.error('Set WHATSAR_KEY dan WHATSAR_SESSION_ID');
  process.exit(1);
}

const to = process.argv[2] ?? '6281234567890';
const text = process.argv[3] ?? 'Halo dari Node.js!';

const res = await fetch(`${baseUrl}/api/v1/messages/send`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': apiKey,
  },
  body: JSON.stringify({ session_id: sessionId, to, text }),
});

console.log('HTTP', res.status);
console.log(JSON.stringify(await res.json(), null, 2));