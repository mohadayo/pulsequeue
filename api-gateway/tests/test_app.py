import pytest
from app import app


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as client:
        yield client


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "healthy"
    assert data["service"] == "api-gateway"
    assert "timestamp" in data


def test_create_job(client):
    resp = client.post("/jobs", json={"task": "send_email", "priority": "high", "payload": {"to": "user@example.com"}})
    assert resp.status_code == 201
    data = resp.get_json()
    assert data["task"] == "send_email"
    assert data["priority"] == "high"
    assert data["status"] == "queued"
    assert "id" in data


def test_create_job_missing_task(client):
    resp = client.post("/jobs", json={"priority": "high"})
    assert resp.status_code == 400
    assert "error" in resp.get_json()


def test_create_job_invalid_priority(client):
    resp = client.post("/jobs", json={"task": "test", "priority": "ultra"})
    assert resp.status_code == 400


def test_list_jobs(client):
    client.post("/jobs", json={"task": "task1"})
    client.post("/jobs", json={"task": "task2"})
    resp = client.get("/jobs")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["total"] >= 2


def test_get_job(client):
    resp = client.post("/jobs", json={"task": "find_me"})
    job_id = resp.get_json()["id"]
    resp = client.get(f"/jobs/{job_id}")
    assert resp.status_code == 200
    assert resp.get_json()["task"] == "find_me"


def test_get_job_not_found(client):
    resp = client.get("/jobs/nonexistent-id")
    assert resp.status_code == 404


def test_cancel_job(client):
    resp = client.post("/jobs", json={"task": "cancel_me"})
    job_id = resp.get_json()["id"]
    resp = client.post(f"/jobs/{job_id}/cancel")
    assert resp.status_code == 200
    assert resp.get_json()["status"] == "cancelled"


def test_cancel_completed_job(client):
    resp = client.post("/jobs", json={"task": "done_job"})
    job_id = resp.get_json()["id"]
    from app import jobs
    jobs[job_id]["status"] = "completed"
    resp = client.post(f"/jobs/{job_id}/cancel")
    assert resp.status_code == 409
