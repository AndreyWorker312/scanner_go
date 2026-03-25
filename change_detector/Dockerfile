# ── Stage 1: install dependencies ──────────────────────────────────────────────
FROM python:3.11-slim AS deps

WORKDIR /build

COPY requirements.txt .

# Install to system site-packages (NOT --user) so the path is predictable
# and works for any user in the final image
RUN pip install --no-cache-dir -r requirements.txt

# ── Stage 2: final minimal image ───────────────────────────────────────────────
FROM python:3.11-slim

WORKDIR /app

# Copy installed packages from the build stage into the system site-packages
# of the final image — available to every user, including non-root
COPY --from=deps /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --from=deps /usr/local/bin /usr/local/bin

# Copy application source
COPY . .

# Non-root user for security
RUN useradd -r -u 1001 -s /sbin/nologin appuser \
    && chown -R appuser /app
USER appuser

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

CMD ["python", "-u", "main.py"]
