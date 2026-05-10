import os
import logging
import uuid
from datetime import datetime
from flask import Flask, jsonify, request

app = Flask(__name__)

logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO"),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("pulsequeue-api")

jobs: dict = {}

WORKER_URL = os.getenv("WORKER_URL", "http://worker-engine:8081")


@app.route("/health")
def health():
    return jsonify({"status": "healthy", "service": "api-gateway", "timestamp": datetime.utcnow().isoformat()})


@app.route("/jobs", methods=["POST"])
def create_job():
    data = request.get_json()
    if not data or "task" not in data:
        logger.warning("Invalid job creation request: missing 'task' field")
        return jsonify({"error": "Field 'task' is required"}), 400

    job_id = str(uuid.uuid4())
    priority = data.get("priority", "normal")
    if priority not in ("low", "normal", "high", "critical"):
        return jsonify({"error": "Invalid priority. Must be one of: low, normal, high, critical"}), 400

    job = {
        "id": job_id,
        "task": data["task"],
        "priority": priority,
        "status": "queued",
        "created_at": datetime.utcnow().isoformat(),
        "payload": data.get("payload", {}),
    }
    jobs[job_id] = job
    logger.info(f"Job created: {job_id} (task={data['task']}, priority={priority})")
    return jsonify(job), 201


@app.route("/jobs", methods=["GET"])
def list_jobs():
    status_filter = request.args.get("status")
    result = list(jobs.values())
    if status_filter:
        result = [j for j in result if j["status"] == status_filter]
    logger.info(f"Listed {len(result)} jobs")
    return jsonify({"jobs": result, "total": len(result)})


@app.route("/jobs/<job_id>", methods=["GET"])
def get_job(job_id):
    job = jobs.get(job_id)
    if not job:
        return jsonify({"error": "Job not found"}), 404
    return jsonify(job)


@app.route("/jobs/<job_id>/cancel", methods=["POST"])
def cancel_job(job_id):
    job = jobs.get(job_id)
    if not job:
        return jsonify({"error": "Job not found"}), 404
    if job["status"] in ("completed", "failed"):
        return jsonify({"error": f"Cannot cancel job in '{job['status']}' state"}), 409
    job["status"] = "cancelled"
    logger.info(f"Job cancelled: {job_id}")
    return jsonify(job)


if __name__ == "__main__":
    port = int(os.getenv("API_PORT", "8080"))
    logger.info(f"Starting PulseQueue API Gateway on port {port}")
    app.run(host="0.0.0.0", port=port)
