# Deployment Guide

This guide covers deploying the go-service-template to production using Docker, Kubernetes, and cloud platforms.

## Table of Contents

1. [Docker Deployment](#docker-deployment)
2. [Kubernetes Deployment](#kubernetes-deployment)
3. [Cloud Platforms](#cloud-platforms)
4. [Environment Configuration](#environment-configuration)
5. [Database Migrations](#database-migrations)
6. [Monitoring & Troubleshooting](#monitoring--troubleshooting)

---

## Docker Deployment

### Building the Image

```bash
# Build Docker image
make docker-build

# Verify image exists
docker image ls | grep go-service-template

# Tag for registry
docker tag go-service-template:latest registry.example.com/go-service-template:v1.2.3
```

### Image Characteristics

**Base:** `debian:bookworm-slim` (development) → `distroless:base-debian12` (production)

The Dockerfile uses multi-stage builds:

1. **Builder stage** — Full Go build environment
2. **Runtime stage** — Distroless image with only the binary

**Benefits of distroless:**
- ✓ Minimal image size (~30MB)
- ✓ No shell or package manager (smaller attack surface)
- ✓ Single-process container (easier to reason about)

### Running in Docker

```bash
# Local development stack (with postgres, redis, prometheus, grafana)
make docker-run

# Or manually
docker run \
  -p 8080:8080 \
  -e APP_DATABASE_DSN=postgres://user:pass@host:5432/db \
  -e APP_REDIS_ADDR=redis:6379 \
  -e APP_JWT_SECRET=your-secret \
  go-service-template:latest
```

### Environment Variables

```bash
# Database
APP_DATABASE_DSN=postgres://user:password@postgres.example.com:5432/service_db?sslmode=require
APP_DATABASE_MAX_OPEN_CONNS=25

# Redis
APP_REDIS_ADDR=redis.example.com:6379
APP_REDIS_PASSWORD=your-password

# Server
APP_SERVER_ADDR=:8080
APP_SERVER_SHUTDOWN_TIMEOUT=30s

# Security
APP_JWT_SECRET=your-very-long-secret-key-minimum-32-chars

# Logging
APP_LOG_LEVEL=info
APP_LOG_DEVELOPMENT=false

# Monitoring
APP_TELEMETRY_ENABLED=true
APP_TELEMETRY_OTLP_ENDPOINT=http://otel-collector:4318
```

### Health Checks

Docker health checks ensure the container is healthy:

```dockerfile
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health/ready || exit 1
```

---

## Kubernetes Deployment

### Using Helm (Recommended)

The repository includes a Helm chart in `deploy/charts/service/`:

```bash
# Install the chart
helm install my-service deploy/charts/service/ \
  --namespace production \
  --create-namespace \
  -f deploy/charts/service/values-prod.yaml

# Upgrade an existing release
helm upgrade my-service deploy/charts/service/ \
  --namespace production \
  -f deploy/charts/service/values-prod.yaml

# Verify installation
kubectl get deployments -n production
kubectl get pods -n production
kubectl logs -n production deployment/my-service
```

### Helm Chart Structure

```
deploy/charts/service/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default chart values
├── values-prod.yaml        # Production overrides
├── values-staging.yaml     # Staging overrides
└── templates/
    ├── deployment.yaml     # Pod + ReplicaSet
    ├── service.yaml        # K8s Service
    ├── ingress.yaml        # HTTP routing
    ├── hpa.yaml            # Horizontal Pod Autoscaler
    ├── configmap.yaml      # Configuration
    ├── serviceaccount.yaml  # RBAC
    └── _helpers.tpl        # Template helpers
```

### Customizing for Production

Create `values-prod.yaml`:

```yaml
# Production deployments
replicaCount: 3

image:
  repository: registry.example.com/go-service-template
  tag: v1.2.3
  pullPolicy: IfNotPresent

# Resource limits
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

# Autoscaling
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

# Ingress
ingress:
  enabled: true
  className: nginx  # Your ingress class
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix

# Environment variables
env:
  APP_LOG_LEVEL: info
  APP_TELEMETRY_ENABLED: "true"
  APP_DATABASE_MAX_OPEN_CONNS: "50"

# Secrets (reference external secret management)
secrets:
  database_dsn: &database_dsn   # Ref to external secret
  redis_password: &redis_pass   # Ref to external secret
  jwt_secret: &jwt_secret       # Ref to external secret
```

### Pod Lifecycle

**Readiness Probe:**
```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

Container only receives traffic when ready.

**Liveness Probe:**
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 20
```

Container is restarted if unhealthy.

**Graceful Shutdown:**
```yaml
terminationGracePeriodSeconds: 30
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 5"]  # Allow connections to drain
```

### Monitoring

**Prometheus scrape config:**
```yaml
scrape_configs:
  - job_name: 'go-service-template'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - production
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: "true"
```

**View metrics:**
```bash
kubectl port-forward -n production svc/prometheus 9090:9090
# Visit http://localhost:9090
```

---

## Cloud Platforms

### AWS ECS Fargate

```bash
# 1. Push image to ECR
aws ecr get-login-password | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com
docker tag go-service-template:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/go-service-template:latest
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/go-service-template:latest

# 2. Create ECS task definition (JSON)
{
  "family": "go-service-template",
  "networkMode": "awsvpc",
  "containers": [
    {
      "name": "app",
      "image": "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/go-service-template:latest",
      "portMappings": [{"containerPort": 8080}],
      "environment": [
        {"name": "APP_LOG_LEVEL", "value": "info"},
        {"name": "APP_DATABASE_DSN", "value": "..."}
      ]
    }
  ],
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512"
}

# 3. Create ECS service
aws ecs create-service \
  --cluster production \
  --service-name go-service-template \
  --task-definition go-service-template:1 \
  --desired-count 3 \
  --launch-type FARGATE
```

### Google Cloud Run

```bash
# Build and push to Artifact Registry
gcloud builds submit --tag gcr.io/$PROJECT_ID/go-service-template

# Deploy
gcloud run deploy go-service-template \
  --image gcr.io/$PROJECT_ID/go-service-template:latest \
  --platform managed \
  --region us-central1 \
  --memory 512M \
  --cpu 1 \
  --set-env-vars="APP_LOG_LEVEL=info" \
  --allow-unauthenticated
```

### Azure Container Instances

```bash
# Push to ACR
az acr build -r $REGISTRY_NAME -f Dockerfile -t go-service-template:latest .

# Deploy
az container create \
  --resource-group myResourceGroup \
  --name go-service-template \
  --image $REGISTRY_NAME.azurecr.io/go-service-template:latest \
  --ports 8080 \
  --environment-variables APP_LOG_LEVEL=info
```

---

## Environment Configuration

### Secrets Management

**Never commit secrets to Git.** Use one of:

#### AWS Secrets Manager
```bash
# Store secret
aws secretsmanager create-secret \
  --name go-service-template/db \
  --secret-string '{"dsn":"postgres://..."}'

# Reference in Kubernetes
env:
  - name: APP_DATABASE_DSN
    valueFrom:
      secretKeyRef:
        name: go-service-template-db
        key: dsn
```

#### Kubernetes Secrets
```bash
# Create secret
kubectl create secret generic go-service-template-secrets \
  -n production \
  --from-literal=APP_JWT_SECRET='...' \
  --from-literal=APP_DATABASE_DSN='...'

# Reference in deployment
env:
  - name: APP_JWT_SECRET
    valueFrom:
      secretKeyRef:
        name: go-service-template-secrets
        key: APP_JWT_SECRET
```

#### HashiCorp Vault
```yaml
# Helm values
vault:
  enabled: true
  role: go-service-template
  path: secret/go-service-template
  secretKeys:
    - APP_JWT_SECRET
    - APP_DATABASE_DSN
    - APP_REDIS_PASSWORD
```

### Configuration Hierarchy

Configuration loads in this order (later overrides earlier):

1. `.env.example` (defaults)
2. `.env` file (if exists)
3. Environment variables
4. Command-line flags (not implemented in this template)

---

## Database Migrations

### Running Migrations

**Before deploying, always run migrations:**

```bash
# Local development
make migrate-up

# In Docker
docker run \
  -e APP_DATABASE_DSN=postgres://user:pass@host:5432/db \
  go-service-template:latest \
  ./scripts/migrate.sh up

# In Kubernetes job
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: migrations
  namespace: production
spec:
  template:
    spec:
      containers:
      - name: migrations
        image: registry.example.com/go-service-template:latest
        command: ["/bin/sh", "-c", "./scripts/migrate.sh up"]
        env:
        - name: APP_DATABASE_DSN
          valueFrom:
            secretKeyRef:
              name: go-service-template-secrets
              key: APP_DATABASE_DSN
      restartPolicy: Never
  backoffLimit: 3
EOF
```

### Zero-Downtime Deployment

For zero-downtime deployments:

1. **Backward-compatible migrations** — Add columns as nullable before using them
2. **Blue-green deployment** — Run migrations, then switch traffic
3. **Canary deployment** — Gradually roll out new version while monitoring

**Example migration sequence:**

```sql
-- Migration 1: Add column (nullable)
ALTER TABLE users ADD COLUMN new_field VARCHAR(255);

-- Deploy: Code reads/writes new_field, falls back to old field
-- ...

-- Migration 2: Drop old column
ALTER TABLE users DROP COLUMN old_field;

-- Deploy: Code only uses new_field
```

---

## Monitoring & Troubleshooting

### Health Check Endpoints

```bash
# Liveness (is the service running?)
curl http://localhost:8080/health/live

# Readiness (is the service ready to handle traffic?)
curl http://localhost:8080/health/ready

# Response (if healthy):
{"status":"ok"}
```

### Viewing Logs

```bash
# Local Docker
docker logs -f <container-id>

# Kubernetes
kubectl logs -n production deployment/go-service-template

# Follow logs in real-time
kubectl logs -n production -f deployment/go-service-template

# Logs from a specific pod
kubectl logs -n production pod/go-service-template-abc123
```

### Debugging

```bash
# Port-forward to local machine
kubectl port-forward -n production svc/go-service-template 8080:8080

# Access service
curl http://localhost:8080/health/ready

# View pod details
kubectl describe pod -n production <pod-name>

# Execute command in pod
kubectl exec -n production -it <pod-name> -- /bin/bash
```

### Common Issues

**Pods not starting:**
```bash
kubectl describe pod -n production <pod-name>
# Check Events section for errors
# Common: "ImagePullBackOff" → Docker image not found
# Common: "CrashLoopBackOff" → Application crashes on startup
```

**Metrics not showing up:**
```bash
# Verify metrics endpoint
kubectl port-forward -n production svc/go-service-template 8080:8080
curl http://localhost:8080/metrics | head -20

# Check Prometheus scrape config
kubectl get configmap -n monitoring prometheus-config -o yaml
```

**Database connection errors:**
```bash
# Verify DATABASE_DSN is set
kubectl exec -n production <pod-name> -- env | grep DATABASE_DSN

# Test connection from pod
kubectl exec -n production <pod-name> -- \
  psql "$APP_DATABASE_DSN" -c "SELECT 1"
```

---

## Deployment Checklist

Before deploying to production:

- [ ] All tests pass (`make test`)
- [ ] Code linting passes (`make lint`)
- [ ] Docker image builds successfully (`make docker-build`)
- [ ] Environment variables documented in `README.md`
- [ ] Secrets stored in secure secret management solution
- [ ] Database migrations tested against production schema
- [ ] Health check endpoints responding correctly
- [ ] Monitoring and alerting configured
- [ ] Rollback plan documented
- [ ] PR reviewed and approved
- [ ] Changelog updated
- [ ] Version tag created (`git tag v1.2.3`)

---

## Rollback Procedure

If something goes wrong after deployment:

```bash
# Helm rollback (fastest)
helm rollback my-service -n production

# Kubernetes previous image
kubectl set image deployment/go-service-template \
  go-service-template=registry.example.com/go-service-template:v1.2.2 \
  -n production

# Monitor rollback
kubectl rollout status deployment/go-service-template -n production

# Verify health
curl http://localhost:8080/health/ready
```

---

## Post-Deployment Validation

```bash
# Check deployment status
kubectl get deployment -n production go-service-template

# Verify pods are healthy
kubectl get pods -n production -l app=go-service-template

# Test endpoints
curl https://api.example.com/health/ready
curl https://api.example.com/metrics | grep http_requests_total

# Check error rates
kubectl logs -n production deployment/go-service-template | grep ERROR
```

---

See [README.md](README.md) for configuration reference and [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for local development setup.
