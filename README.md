# Production-Ready Kubernetes Microservices Platform (SRE Project)

A highly available, secure, and observable microservices platform built with Kubernetes, demonstrating real-world Site Reliability Engineering (SRE) practices.

## Overview

This project showcases the design and deployment of a **multi-language microservices architecture** on Kubernetes. It focuses on resilience, scalability, security, and observability — key principles in modern cloud-native environments.

### Architecture

- **Node.js API Service** — Main backend API
- **Go Auth Service** — Authentication & authorization service
- **Python Image Service** — Handles image processing and storage

All services are containerized and orchestrated using Kubernetes.

## Key Features

### Containerization & Orchestration
- Docker containerization for each microservice
- Multi-stage Docker builds for optimized images
- Private Docker registry integration

### Kubernetes Implementation
- **Deployments** with proper resource requests & limits
- **Services** (ClusterIP + LoadBalancer)
- **Ingress** with TLS termination
- **Network Policies** for pod-to-pod security
- **Horizontal Pod Autoscaler (HPA)** for automatic scaling
- **Liveness & Readiness Probes** for health checking
- **PodDisruptionBudgets** for high availability

### Observability & Monitoring
- **Prometheus** for metrics collection
- **Grafana** dashboards with custom metrics
- **Alertmanager** for intelligent alerting
- Centralized logging

### Security & Resilience
- Secrets management
- Network segmentation
- TLS encryption
- Simulated failure scenarios:
  - Database outage
  - Service crashes
  - Traffic spikes
- Chaos engineering testing

## Technologies Used

- **Orchestration**: Kubernetes (K8s)
- **Containerization**: Docker
- **Languages**: Node.js, Go, Python
- **Monitoring**: Prometheus, Grafana, Alertmanager
- **Ingress & Security**: NGINX Ingress, Network Policies
- **Others**: Helm (optional), GitOps practices

## Project Goals

- Demonstrate production-grade Kubernetes deployment patterns
- Implement SRE best practices (observability, reliability, automation)
- Showcase multi-language microservices communication
- Validate system resilience under failure conditions

## How to Run (Local / Minikube)

```bash
# Clone the repo
git clone <repository-url>
cd kubernetes-microservices-sre

# Apply Kubernetes manifests
kubectl apply -f manifests/
