# Deployment Guide

## Local Development

See [QUICKSTART.md](./QUICKSTART.md) for local setup instructions.

## Production Deployment

### Prerequisites

- Kubernetes cluster (GKE, EKS, AKS, or self-managed)
- kubectl configured
- Helm 3.x installed
- Docker registry access

### Option 1: Docker Swarm

#### Initialize Swarm

```bash
docker swarm init
```

#### Deploy Stack

```bash
docker stack deploy -c docker-compose.yml reco-engine
```

#### Scale Services

```bash
docker service scale reco-engine_ingest=3
docker service scale reco-engine_api=3
docker service scale reco-engine_processor=2
```

### Option 2: Kubernetes

#### 1. Build and Push Images

```bash
# Build images
docker build -t your-registry/reco-ingest:latest --build-arg SERVICE=ingest .
docker build -t your-registry/reco-processor:latest --build-arg SERVICE=processor .
docker build -t your-registry/reco-api:latest --build-arg SERVICE=api .

# Push to registry
docker push your-registry/reco-ingest:latest
docker push your-registry/reco-processor:latest
docker push your-registry/reco-api:latest
```

#### 2. Create Namespace

```bash
kubectl create namespace reco-engine
```

#### 3. Deploy Infrastructure

```yaml
# postgres-deployment.yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: reco-engine
spec:
  ports:
  - port: 5432
  selector:
    app: postgres
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: reco-engine
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_DB
          value: reco
        - name: POSTGRES_USER
          value: reco
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 20Gi
```

```bash
kubectl apply -f postgres-deployment.yaml
```

#### 4. Deploy Services

```yaml
# ingest-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ingest
  namespace: reco-engine
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ingest
  template:
    metadata:
      labels:
        app: ingest
    spec:
      containers:
      - name: ingest
        image: your-registry/reco-ingest:latest
        ports:
        - containerPort: 8080
        env:
        - name: RECO_KAFKA_BROKERS
          value: "kafka:9092"
        - name: RECO_POSTGRES_HOST
          value: "postgres"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: ingest
  namespace: reco-engine
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: ingest
```

```bash
kubectl apply -f ingest-deployment.yaml
kubectl apply -f api-deployment.yaml
kubectl apply -f processor-deployment.yaml
```

#### 5. Configure Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: reco-engine
  namespace: reco-engine
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.yourdomain.com
    secretName: reco-tls
  rules:
  - host: api.yourdomain.com
    http:
      paths:
      - path: /events
        pathType: Prefix
        backend:
          service:
            name: ingest
            port:
              number: 80
      - path: /recommendations
        pathType: Prefix
        backend:
          service:
            name: api
            port:
              number: 80
```

### Option 3: Cloud Managed Services

#### AWS

- **ECS/Fargate** for containers
- **RDS PostgreSQL** for database
- **ElastiCache Redis** for caching
- **MSK (Kafka)** for streaming
- **ALB** for load balancing
- **CloudWatch** for monitoring

#### GCP

- **GKE** for Kubernetes
- **Cloud SQL** for PostgreSQL
- **Memorystore Redis** for caching
- **Pub/Sub** (alternative to Kafka)
- **Cloud Load Balancing**
- **Cloud Monitoring**

#### Azure

- **AKS** for Kubernetes
- **Azure Database for PostgreSQL**
- **Azure Cache for Redis**
- **Event Hubs** (Kafka-compatible)
- **Application Gateway**
- **Azure Monitor**

### Scaling Guidelines

#### Horizontal Scaling

**Ingest Service:**
- Scale based on requests per second
- Target: CPU < 70%, Memory < 80%
- Recommended: 3-10 replicas

**API Service:**
- Scale based on concurrent users
- Target: P99 latency < 100ms
- Recommended: 3-10 replicas

**Processor Service:**
- Scale based on Kafka consumer lag
- One replica per Kafka partition
- Recommended: 3-5 replicas

#### Vertical Scaling

**Development:**
- CPU: 0.1-0.5 cores
- Memory: 128-512 MB

**Production:**
- CPU: 0.5-2 cores
- Memory: 512 MB - 2 GB

### Database Configuration

#### PostgreSQL

**Connection Pool:**
```yaml
max_connections: 100
max_open_conns: 25
max_idle_conns: 5
```

**Performance Tuning:**
```sql
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
```

#### Redis

**Memory:**
- Development: 512 MB
- Production: 2-8 GB

**Configuration:**
```
maxmemory 2gb
maxmemory-policy allkeys-lru
appendonly yes
appendfsync everysec
```

### Monitoring & Alerting

#### Prometheus Alerts

```yaml
groups:
- name: reco-engine
  rules:
  - alert: HighErrorRate
    expr: rate(event_processing_errors_total[5m]) > 0.05
    for: 5m
    annotations:
      summary: High error rate detected
  
  - alert: HighLatency
    expr: histogram_quantile(0.99, rate(recommendation_latency_seconds_bucket[5m])) > 0.5
    for: 5m
    annotations:
      summary: High recommendation latency
  
  - alert: KafkaConsumerLag
    expr: kafka_consumer_lag > 10000
    for: 5m
    annotations:
      summary: High Kafka consumer lag
```

#### Grafana Dashboards

Import dashboards for:
- Service metrics (requests, latency, errors)
- Infrastructure metrics (CPU, memory, disk)
- Business metrics (recommendations served, cache hit ratio)

### Security Hardening

#### Network

- Use VPC/Private networks
- Restrict ingress to load balancer only
- Enable TLS for all services
- Use network policies in Kubernetes

#### Authentication

```yaml
# Add API key middleware
middlewares:
  - apiKey:
      secretName: api-keys
```

#### Secrets Management

```bash
# Create secrets
kubectl create secret generic postgres-secret \
  --from-literal=password='your-secure-password'

kubectl create secret generic kafka-credentials \
  --from-file=ca.crt \
  --from-file=client.crt \
  --from-file=client.key
```

### Backup & Disaster Recovery

#### PostgreSQL Backup

```bash
# Automated backup with pg_dump
0 2 * * * pg_dump -U reco -d reco | gzip > /backup/reco-$(date +\%Y\%m\%d).sql.gz
```

#### Redis Backup

```bash
# Enable RDB snapshots
redis-cli CONFIG SET save "900 1 300 10 60 10000"
```

#### Kafka

- Enable replication factor: 3
- Enable log retention: 7 days
- Regular topic backups

### CI/CD Pipeline

#### GitHub Actions Example

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Build images
      run: |
        docker build -t ${{ secrets.REGISTRY }}/reco-ingest:${{ github.sha }} --build-arg SERVICE=ingest .
        docker build -t ${{ secrets.REGISTRY }}/reco-api:${{ github.sha }} --build-arg SERVICE=api .
        docker build -t ${{ secrets.REGISTRY }}/reco-processor:${{ github.sha }} --build-arg SERVICE=processor .
    
    - name: Push images
      run: |
        echo ${{ secrets.REGISTRY_PASSWORD }} | docker login -u ${{ secrets.REGISTRY_USER }} --password-stdin
        docker push ${{ secrets.REGISTRY }}/reco-ingest:${{ github.sha }}
        docker push ${{ secrets.REGISTRY }}/reco-api:${{ github.sha }}
        docker push ${{ secrets.REGISTRY }}/reco-processor:${{ github.sha }}
    
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/ingest ingest=${{ secrets.REGISTRY }}/reco-ingest:${{ github.sha }}
        kubectl set image deployment/api api=${{ secrets.REGISTRY }}/reco-api:${{ github.sha }}
        kubectl set image deployment/processor processor=${{ secrets.REGISTRY }}/reco-processor:${{ github.sha }}
```

### Cost Optimization

1. **Right-size instances** based on actual usage
2. **Use spot instances** for non-critical workloads
3. **Enable autoscaling** to scale down during low traffic
4. **Use reserved instances** for predictable workloads
5. **Implement caching** to reduce database load
6. **Monitor costs** with cloud provider tools
