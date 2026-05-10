import express, { Request, Response } from "express";

const app = express();
app.use(express.json());

const PORT = process.env.DASHBOARD_PORT || "8082";
const API_GATEWAY_URL = process.env.API_GATEWAY_URL || "http://api-gateway:8080";
const WORKER_URL = process.env.WORKER_URL || "http://worker-engine:8081";

interface ServiceStatus {
  name: string;
  url: string;
  status: "up" | "down" | "unknown";
  lastChecked: string;
}

const serviceStatuses: ServiceStatus[] = [
  { name: "api-gateway", url: API_GATEWAY_URL, status: "unknown", lastChecked: "" },
  { name: "worker-engine", url: WORKER_URL, status: "unknown", lastChecked: "" },
];

app.get("/health", (_req: Request, res: Response) => {
  res.json({
    status: "healthy",
    service: "dashboard",
    timestamp: new Date().toISOString(),
  });
});

app.get("/api/services", (_req: Request, res: Response) => {
  res.json({ services: serviceStatuses });
});

app.get("/api/overview", async (_req: Request, res: Response) => {
  const overview = {
    platform: "PulseQueue",
    version: "1.0.0",
    services: serviceStatuses.length,
    uptime: process.uptime(),
    timestamp: new Date().toISOString(),
  };
  res.json(overview);
});

app.get("/api/config", (_req: Request, res: Response) => {
  res.json({
    apiGatewayUrl: API_GATEWAY_URL,
    workerUrl: WORKER_URL,
    dashboardPort: PORT,
    environment: process.env.NODE_ENV || "development",
  });
});

export { app };

if (require.main === module) {
  app.listen(parseInt(PORT), "0.0.0.0", () => {
    console.log(`[dashboard] PulseQueue Dashboard running on port ${PORT}`);
    console.log(`[dashboard] API Gateway URL: ${API_GATEWAY_URL}`);
    console.log(`[dashboard] Worker URL: ${WORKER_URL}`);
  });
}
