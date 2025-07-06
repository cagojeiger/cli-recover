# CLI-Restore Roadmap

## Phase 1: Foundation (v0.1.0) ✅
- [x] Context Engineering 구조 설정
- [x] 기본 프로젝트 구조
- [x] `--version` 명령어 구현
- [x] 빌드 시스템 (Makefile)
- [x] GitHub Actions 자동 릴리즈

## Phase 2: TUI Framework (v0.2.0) - Current
- [x] Bubble Tea 프레임워크 선택
- [x] 표준 레이아웃 설계 (Header/Main/Preview/Footer)
- [x] 명령어 패턴 확립 (`[action] [target] [options]`)
- [ ] 기본 TUI 구현
- [ ] 네비게이션 시스템
- [ ] Pod 파일시스템 백업

## Phase 3: Core Backup Features (v0.3.0)
- [ ] MongoDB 백업 지원 (Bitnami)
- [ ] MinIO 백업 지원 (mc 자동 처리)
- [ ] PostgreSQL 백업 지원
- [ ] 크기 기반 백업 전략 자동 선택
- [ ] Port forward 자동 관리
- [ ] 진행률 실시간 표시

## Phase 4: Offline Support (v0.4.0)
- [ ] Go embed 기반 바이너리 임베딩
- [ ] mc 바이너리 자동 주입
- [ ] 오프라인 빌드 변형 (minimal/standard/full)
- [ ] 바이너리 보안 검증
- [ ] 자동 정리 보장

## Phase 5: Advanced Features (v0.5.0)
- [ ] 복원 기능 완성
- [ ] 백업 검증 (verify)
- [ ] 백업 히스토리 관리
- [ ] 기본 스케줄링 지원
- [ ] 에러 복구 메커니즘

## Phase 6: Enterprise Ready (v1.0.0)
- [ ] Multi-cluster 지원
- [ ] Cloud storage 통합 (S3, GCS, Azure)
- [ ] 암호화 지원
- [ ] RBAC 통합
- [ ] Prometheus 메트릭
- [ ] Webhook 알림

## Phase 7: Extended Services (v1.1.0)
- [ ] Redis 백업 지원
- [ ] MySQL/MariaDB 지원
- [ ] Elasticsearch 지원
- [ ] 증분 백업
- [ ] 병렬 백업 최적화

## Phase 8: Platform Features (v1.2.0)
- [ ] Web Dashboard
- [ ] Helm Chart
- [ ] Kubernetes Operator
- [ ] Plugin 시스템
- [ ] Community templates

## Current Status
- **Released**: v0.1.0 (Foundation complete)
- **In Progress**: v0.2.0 (TUI Framework)
- **Next Focus**: Bubble Tea 기반 TUI 구현