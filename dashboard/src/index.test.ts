import request from "supertest";
import { app } from "./index";

describe("Dashboard API", () => {
  describe("GET /health", () => {
    it("should return healthy status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("healthy");
      expect(res.body.service).toBe("dashboard");
      expect(res.body.timestamp).toBeDefined();
    });
  });

  describe("GET /api/services", () => {
    it("should return service list", async () => {
      const res = await request(app).get("/api/services");
      expect(res.status).toBe(200);
      expect(res.body.services).toBeInstanceOf(Array);
      expect(res.body.services.length).toBe(2);
      expect(res.body.services[0].name).toBe("api-gateway");
      expect(res.body.services[1].name).toBe("worker-engine");
    });
  });

  describe("GET /api/overview", () => {
    it("should return platform overview", async () => {
      const res = await request(app).get("/api/overview");
      expect(res.status).toBe(200);
      expect(res.body.platform).toBe("PulseQueue");
      expect(res.body.version).toBe("1.0.0");
      expect(res.body.services).toBe(2);
      expect(res.body.uptime).toBeGreaterThan(0);
    });
  });

  describe("GET /api/config", () => {
    it("should return configuration", async () => {
      const res = await request(app).get("/api/config");
      expect(res.status).toBe(200);
      expect(res.body.apiGatewayUrl).toBeDefined();
      expect(res.body.workerUrl).toBeDefined();
      expect(res.body.dashboardPort).toBeDefined();
    });
  });
});
