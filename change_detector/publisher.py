"""
RabbitMQ publisher for change events.

Publishes detected ChangeEvent dicts to the `change_events` queue.
The backend Go service consumes from this queue and broadcasts to WebSocket clients.
"""
import json
import logging
import time

import pika
import pika.exceptions

logger = logging.getLogger(__name__)


class ChangeEventPublisher:
    QUEUE = "change_events"

    def __init__(self, rabbitmq_url: str):
        self._url    = rabbitmq_url
        self._conn   = None
        self._channel = None

    # ── Connection lifecycle ──────────────────────────────────────────────────

    def connect(self, max_attempts: int = 10, delay: int = 5) -> None:
        for attempt in range(1, max_attempts + 1):
            try:
                params = pika.URLParameters(self._url)
                params.heartbeat = 60
                params.blocked_connection_timeout = 30
                self._conn    = pika.BlockingConnection(params)
                self._channel = self._conn.channel()
                self._channel.queue_declare(
                    queue=self.QUEUE,
                    durable=True,   # survives broker restart
                )
                logger.info("RabbitMQ publisher connected (attempt %d/%d), queue: %s",
                            attempt, max_attempts, self.QUEUE)
                return
            except Exception as exc:
                logger.warning("RabbitMQ not ready (attempt %d/%d): %s", attempt, max_attempts, exc)
                if attempt < max_attempts:
                    time.sleep(delay)

        raise RuntimeError(f"Failed to connect to RabbitMQ after {max_attempts} attempts")

    def close(self) -> None:
        try:
            if self._conn and not self._conn.is_closed:
                self._conn.close()
        except Exception as exc:
            logger.debug("Error closing RabbitMQ connection: %s", exc)

    # ── Publishing ────────────────────────────────────────────────────────────

    def publish(self, event: dict) -> bool:
        """
        Publish a single ChangeEvent dict to the change_events queue.
        Returns True on success, False on failure (caller should decide whether to retry).
        """
        try:
            body = json.dumps(event, default=str)
            self._channel.basic_publish(
                exchange="",
                routing_key=self.QUEUE,
                body=body,
                properties=pika.BasicProperties(
                    content_type="application/json",
                    delivery_mode=2,   # persistent message
                ),
            )
            return True
        except (pika.exceptions.AMQPError, AttributeError) as exc:
            logger.error("Failed to publish change event: %s — reconnecting…", exc)
            self._try_reconnect()
            return False

    # ── Internals ─────────────────────────────────────────────────────────────

    def _try_reconnect(self) -> None:
        try:
            self.close()
        except Exception:
            pass
        try:
            self.connect(max_attempts=3, delay=2)
        except Exception as exc:
            logger.error("Reconnect failed: %s", exc)

