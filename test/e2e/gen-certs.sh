#!/bin/bash
set -e

# gen-certs.sh - Generate self-signed TLS certificates for mTLS testing
# Generates CA, server, and client certificates with SANs for Docker Compose services

CERT_DIR="$(dirname "$0")/certs"
VALIDITY_DAYS=365
KEY_SIZE=2048

# Cleanup existing certificates
echo "[*] Cleaning up existing certificates..."
rm -rf "$CERT_DIR"
mkdir -p "$CERT_DIR"

# 1. Generate CA private key
echo "[*] Generating CA private key..."
openssl genrsa -out "$CERT_DIR/ca.key" $KEY_SIZE 2>/dev/null

# 2. Generate CA certificate
echo "[*] Generating CA certificate..."
openssl req -new -x509 \
  -days $VALIDITY_DAYS \
  -key "$CERT_DIR/ca.key" \
  -out "$CERT_DIR/ca.crt" \
  -subj "/CN=HybridGrid-CA/O=HybridGrid/C=US" 2>/dev/null

# 3. Generate server private key
echo "[*] Generating server private key..."
openssl genrsa -out "$CERT_DIR/server.key" $KEY_SIZE 2>/dev/null

# 4. Generate server certificate signing request with SANs
echo "[*] Generating server certificate signing request..."
openssl req -new \
  -key "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.csr" \
  -subj "/CN=localhost/O=HybridGrid/C=US" \
  -addext "subjectAltName=DNS:localhost,DNS:coordinator,DNS:worker-1,DNS:worker-2" 2>/dev/null

# 5. Sign server certificate with CA
echo "[*] Signing server certificate with CA..."
openssl x509 -req \
  -in "$CERT_DIR/server.csr" \
  -CA "$CERT_DIR/ca.crt" \
  -CAkey "$CERT_DIR/ca.key" \
  -CAcreateserial \
  -out "$CERT_DIR/server.crt" \
  -days $VALIDITY_DAYS \
  -copy_extensions copy 2>/dev/null

# 6. Generate client private key
echo "[*] Generating client private key..."
openssl genrsa -out "$CERT_DIR/client.key" $KEY_SIZE 2>/dev/null

# 7. Generate client certificate signing request
echo "[*] Generating client certificate signing request..."
openssl req -new \
  -key "$CERT_DIR/client.key" \
  -out "$CERT_DIR/client.csr" \
  -subj "/CN=hybridgrid-client/O=HybridGrid/C=US" 2>/dev/null

# 8. Sign client certificate with CA
echo "[*] Signing client certificate with CA..."
openssl x509 -req \
  -in "$CERT_DIR/client.csr" \
  -CA "$CERT_DIR/ca.crt" \
  -CAkey "$CERT_DIR/ca.key" \
  -CAcreateserial \
  -out "$CERT_DIR/client.crt" \
  -days $VALIDITY_DAYS 2>/dev/null

# 9. Set proper file permissions
echo "[*] Setting file permissions..."
chmod 600 "$CERT_DIR/ca.key" "$CERT_DIR/server.key" "$CERT_DIR/client.key"
chmod 644 "$CERT_DIR/ca.crt" "$CERT_DIR/server.crt" "$CERT_DIR/client.crt"

# 10. Clean up CSR files
echo "[*] Cleaning up certificate signing requests..."
rm -f "$CERT_DIR/"*.csr "$CERT_DIR/"*.srl

echo "[✓] Certificates generated successfully!"
echo "[✓] Output directory: $CERT_DIR"
ls -lh "$CERT_DIR"
