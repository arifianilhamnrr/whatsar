<?php
/**
 * Kirim dokumen/PDF dari PHP
 * Usage: php send-document.php 6281234567890 /path/file.pdf
 */

$url     = getenv('WHATSAR_URL') ?: 'http://127.0.0.1:8080';
$apiKey  = getenv('WHATSAR_KEY') ?: die("WHATSAR_KEY belum diset\n");
$session = getenv('WHATSAR_SESSION_ID') ?: die("WHATSAR_SESSION_ID belum diset\n");

$nomor = $argv[1] ?? die("Usage: php send-document.php NOMOR PATH\n");
$path  = $argv[2] ?? die("Usage: php send-document.php NOMOR PATH\n");

if (!is_readable($path)) {
    die("File tidak ditemukan: {$path}\n");
}

$payload = json_encode([
    'session_id'      => $session,
    'to'              => $nomor,
    'type'            => 'document',
    'document_base64' => base64_encode(file_get_contents($path)),
    'filename'        => basename($path),
    'mimetype'        => mime_content_type($path) ?: 'application/octet-stream',
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
    CURLOPT_TIMEOUT        => 120,
]);

$response = curl_exec($ch);
echo "HTTP " . curl_getinfo($ch, CURLINFO_HTTP_CODE) . "\n";
echo $response . "\n";
curl_close($ch);