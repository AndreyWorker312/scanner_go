"""
Change Detector — entry point.

Connects to MongoDB (for scan history + change-event dedup) and RabbitMQ
(to publish detected changes to the backend).

Flow:
  MongoDB scan history → detect changes → save to MongoDB change_events
                                        → publish to RabbitMQ change_events queue
  Backend consumes change_events queue → broadcasts via WebSocket to all clients.
"""
import os
import sys
import time
import signal
import logging

from detector  import ChangeDetector
from publisher import ChangeEventPublisher

# ─── Logging ──────────────────────────────────────────────────────────────────
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
    stream=sys.stdout,
    force=True,   # reset any handlers added before this call (e.g. by pymongo)
)
# pymongo and pika are chatty at DEBUG — silence them
logging.getLogger("pymongo").setLevel(logging.WARNING)
logging.getLogger("pika").setLevel(logging.WARNING)
logger = logging.getLogger("change_detector")

# ─── Config from environment ──────────────────────────────────────────────────
MONGO_URI    = os.getenv("MONGODB_URI",       "mongodb://mongodb:27017")
MONGO_DB     = os.getenv("MONGODB_DATABASE",  "network_scanner")
RABBITMQ_URL = os.getenv("RABBITMQ_URL",      "amqp://guest:guest@rabbitmq:5672/")
INTERVAL     = int(os.getenv("DETECTION_INTERVAL", "30"))

# ─── Graceful shutdown ────────────────────────────────────────────────────────
_running = True


def _handle_signal(signum, frame):          # noqa: ANN001
    global _running
    logger.info("Shutdown signal received (%s)", signum)
    _running = False


def main() -> None:
    global _running

    signal.signal(signal.SIGINT,  _handle_signal)
    signal.signal(signal.SIGTERM, _handle_signal)

    logger.info(
        "Change Detector starting | MongoDB: %s | DB: %s | RabbitMQ: %s | Interval: %ds",
        MONGO_URI, MONGO_DB, RABBITMQ_URL, INTERVAL,
    )

    # ── Wait for MongoDB ─────────────────────────────────────────────────────
    detector: ChangeDetector | None = None
    for attempt in range(1, 11):
        try:
            detector = ChangeDetector(MONGO_URI, MONGO_DB)
            detector.db.command("ping")
            logger.info("MongoDB connected (attempt %d/10)", attempt)
            break
        except Exception as exc:
            logger.warning("MongoDB not ready (attempt %d/10): %s", attempt, exc)
            if detector:
                detector.close()
                detector = None
            time.sleep(6)

    if detector is None:
        logger.error("Could not connect to MongoDB after 10 attempts. Exiting.")
        sys.exit(1)

    # ── Ensure indexes ───────────────────────────────────────────────────────
    try:
        detector.ensure_indexes()
    except Exception as exc:
        logger.warning("Index creation warning (non-fatal): %s", exc)

    # ── Connect to RabbitMQ ───────────────────────────────────────────────────
    publisher = ChangeEventPublisher(RABBITMQ_URL)
    try:
        publisher.connect(max_attempts=10, delay=6)
    except RuntimeError as exc:
        logger.error("Could not connect to RabbitMQ: %s. Exiting.", exc)
        detector.close()
        sys.exit(1)

    # ── Detection loop ───────────────────────────────────────────────────────
    logger.info("Detection loop started. Running every %d seconds.", INTERVAL)
    while _running:
        try:
            new_events = detector.run_detection()
            # Publish each newly saved event to RabbitMQ
            for event in new_events:
                publisher.publish(event)
        except Exception as exc:
            logger.error("Unhandled error in detection cycle: %s", exc, exc_info=True)

        # Sleep in 1-second chunks to stay responsive to SIGTERM
        elapsed = 0
        while elapsed < INTERVAL and _running:
            time.sleep(1)
            elapsed += 1

    logger.info("Change Detector stopped gracefully.")
    publisher.close()
    detector.close()


if __name__ == "__main__":
    main()

