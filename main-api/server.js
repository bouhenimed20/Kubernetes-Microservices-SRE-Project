const express = require("express");
const axios = require("axios");
const client = require("prom-client");

const app = express();
app.use(express.json());


const PORT = process.env.PORT || 3000;

const AUTH_SERVICE_URL =
  process.env.AUTH_URL ||
  "http://auth-service.prod-auth.svc.cluster.local:80";

const IMAGE_SERVICE_URL =
  process.env.IMAGE_URL ||
  "http://image-service.prod-image.svc.cluster.local:80";


client.collectDefaultMetrics();

const httpRequestDurationMs = new client.Histogram({
  name: "http_request_duration_ms",
  help: "Duration of HTTP requests in ms",
  labelNames: ["method", "route", "status_code"],
  buckets: [5, 10, 25, 50, 100, 250, 500, 1000, 2000],
});

app.use((req, res, next) => {
  const end = httpRequestDurationMs.startTimer({
    method: req.method,
    route: req.path,
  });
  res.on("finish", () => {
    end({ status_code: res.statusCode });
  });
  next();
});

app.get("/metrics", async (_req, res) => {
  res.set("Content-Type", client.register.contentType);
  res.end(await client.register.metrics());
});

app.get("/health", (_req, res) => res.json({ status: "ok" }));
app.get("/health/live", (_req, res) => res.json({ status: "live" }));
app.get("/health/ready", (_req, res) => res.json({ status: "ready" }));

app.get("/", (_req, res) => {
  res.json({
    service: "main-api",
    authService: AUTH_SERVICE_URL,
    imageService: IMAGE_SERVICE_URL,
  });
});

app.post("/auth/validate", async (req, res) => {
  try {
    const r = await axios.post(`${AUTH_SERVICE_URL}/validate`, req.body, {
      timeout: 3000,
    });
    res.status(r.status).json(r.data);
  } catch (err) {
    res.status(502).json({
      error: "auth-service unavailable",
      details: err.message,
    });
  }
});

app.post("/images/upload", async (req, res) => {
  try {
    const r = await axios.post(`${IMAGE_SERVICE_URL}/upload`, req.body, {
      timeout: 5000,
    });
    res.status(r.status).json(r.data);
  } catch (err) {
    res.status(502).json({
      error: "image-service unavailable",
      details: err.message,
    });
  }
});

app.listen(PORT, () => {
  console.log(`main-api running on port ${PORT}`);
});
