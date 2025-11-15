# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Offline training pipeline for embeddings
- ANN integration (Milvus/Faiss)
- A/B testing framework
- Advanced ranking algorithms
- User authentication and authorization
- GDPR compliance features

## [1.0.0] - 2025-11-15

### Added
- Initial release of Distributed Real-Time Recommendation Engine
- Event ingestion service with Kafka integration
- Real-time stream processor for feature aggregation
- Recommendation API with personalized and popular endpoints
- Redis-backed feature store for low-latency access
- PostgreSQL for metadata and event storage
- Docker Compose setup for local development
- Prometheus metrics and monitoring
- Grafana dashboards support
- Load testing scripts with k6
- Comprehensive documentation (Quick Start, API, Architecture)
- Unit tests for core components

### Features
- **Real-time Processing**: Sub-second event processing latency
- **Hybrid Recommendations**: Combines co-view, popularity, and embedding signals
- **Scalable Architecture**: Microservices design with horizontal scaling
- **Caching**: 5-minute TTL for recommendation results
- **Monitoring**: Full observability with Prometheus and OpenTelemetry
- **Production-ready**: Docker containerization and deployment guides

### Algorithms
- Co-view matrix (session-based item affinity)
- Weighted popularity scoring (VIEW=1, CLICK=3, CART=5, PURCHASE=10)
- Session-based recent items tracking
- Multi-signal scoring with configurable weights

## [0.1.0] - 2025-11-01

### Added
- Project initialization
- Basic service structure
- Development environment setup
