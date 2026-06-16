<?php
/**
 * Contoh kirim notif WhatsApp dari Laragon / PHP
 *
 * .env project:
 *   WHATSAR_URL=http://127.0.0.1:8080
 *   WHATSAR_KEY=your-api-key
 *   WHATSAR_SESSION_ID=uuid-dari-dashboard
 */

$url     = getenv('WHATSAR_URL') ?: 'http://127.0.0.1:8080';
$apiKey  = getenv('WHATSAR_KEY') ?: die("WHATSAR_KEY belum diset\n");
$session = getenv('WHATSAR_SESSION_ID') ?: die("WHATSAR_SESSION_ID belum diset\n");

$nomor = $argv[1] ?? '6281234567890';
$pesan = $argv[2] ?? 'Halo dari Laragon!';

$payload = json_encode([
    'session_id' => $session,
    'to'         => $nomor,
    'text'       => $pesan,
]);

$ch = curl_init("{$url}/api/v1/messages/send");
curl_setopt_array($ch, [
    CURLOPT_POST           => true,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_HTTPHEADER     => [
        'Content-Type: application/json',
        "X-API-Key: {$apiKey}",
    ],
    CURLOPT_POSTFIELDS     => $payload,
    CURLOPT_TIMEOUT        => 30,
]);

$response = curl_exec($ch);
$code     = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

echo "HTTP {$code}\n";
echo $response . "\n";