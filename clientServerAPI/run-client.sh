#!/bin/sh

# Wait for server to be ready
echo "Waiting for server to be ready..."
sleep 5

# Run the client once
echo "Running client to fetch exchange rate..."
./client

echo "Exchange rate fetched successfully. Check /data/cotacao.txt for the result."
