#!/bin/bash
set -e

# Get the mapped host IP
HOST_IP=$(hostname -I | awk '{print $1}')


# Get the mapped host ports for 9092 and 9093 from environment or default
PORT_9092="${KAFKA_PORT_9092:-9092}"
PORT_9093="${KAFKA_PORT_9093:-9093}"

# Set listeners and advertised.listeners for both ports
export KAFKA_LISTENERS="PLAINTEXT://${HOST_IP}:${PORT_9092},CONTROLLER://${HOST_IP}:${PORT_9093}"
export KAFKA_ADVERTISED_LISTENERS="PLAINTEXT://${HOST_IP}:${PORT_9092},CONTROLLER://${HOST_IP}:${PORT_9093}"

# Print for debugging
echo "Starting Kafka with KAFKA_LISTENERS=${KAFKA_LISTENERS}"
echo "Starting Kafka with KAFKA_ADVERTISED_LISTENERS=${KAFKA_ADVERTISED_LISTENERS}"

# Start Kafka
exec /etc/confluent/docker/run
