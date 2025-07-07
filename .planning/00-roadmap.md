# CLI-First 개발 로드맵

## 전략 전환
- TUI 중심 → CLI 우선 개발
- 이유: 동작하는 CLI 백업이 이미 구현되어 있고, 더 실용적
- 원칙: "Make it work → Make it right → Make it pretty"

## 현재 상태
- filesystem 백업 CLI 동작 중
- kubectl exec + tar 통합 완료
- 진행률 모니터링 구현
- 크기 추정 & ETA 계산 가능

## Phase 1: CLI 핵심 완성 (1주)
**복잡도**: 20/100 ✅
**목표**: 모든 백업 타입을 CLI로 실행 가능하게

### 작업 항목
- [ ] CLI 명령 체계 표준화
- [ ] filesystem provider 안정화
- [ ] minio provider 구현
- [ ] mongodb provider 구현
- [ ] 공통 진행률 처리
- [ ] 에러 처리 통일

### 명령 구조
```bash
cli-recover backup <type> <pod> <path> [options]
cli-recover restore <type> <pod> <backup-file> [options]
cli-recover list backups
cli-recover status <job-id>
```

### 성공 지표
- 3가지 백업 타입 모두 동작
- 일관된 진행률/에러 표시
- 기존 filesystem 백업 계속 동작

## Phase 2: 아키텍처 정리 (1주)
**복잡도**: 35/100 ✅
**목표**: 플러그인 패턴으로 확장 가능한 구조 구축

### 작업 항목
- [ ] BackupProvider 인터페이스 정의
- [ ] Provider 레지스트리 구현
- [ ] 도메인 레이어 분리
- [ ] Infrastructure 레이어 구성
- [ ] 의존성 주입 적용
- [ ] 단위 테스트 작성

### 핵심 인터페이스
```go
type BackupProvider interface {
    Name() string
    Execute(ctx context.Context, opts BackupOptions) error
    EstimateSize(opts BackupOptions) (int64, error)
    StreamProgress() <-chan Progress
}
```

### 디렉토리 구조
```
cmd/cli-recover/
├── main.go
├── commands/
└── handlers/

internal/
├── domain/
│   └── backup/
├── application/
│   └── services/
├── infrastructure/
│   ├── kubernetes/
│   └── providers/
└── presentation/
    └── cli/
```

### 성공 지표
- 새 provider 추가 < 200 LOC
- 테스트 커버리지 > 60%
- 모든 의존성 인터페이스화

## Phase 3: CLI 고도화 (2주)
**복잡도**: 40/100 ⚠️
**목표**: 프로덕션 레벨 CLI 도구 완성

### 작업 항목
- [ ] restore 명령 구현
- [ ] list 명령 (백업 목록)
- [ ] status 명령 (작업 상태)
- [ ] 설정 파일 지원
- [ ] 로깅 시스템 구축
- [ ] 병렬 백업 지원

### 추가 기능
- 백업 메타데이터 저장
- 압축 옵션
- 암호화 지원
- 대역폭 제한
- 재시도 로직

### 성공 지표
- 모든 명령 완전 구현
- 설정 파일로 관리 가능
- 로그 레벨 조절 가능
- 에러 시 자동 재시도

## Phase 4: TUI 래핑 (2주)
**복잡도**: 45/100 ⚠️
**목표**: CLI 위에 사용자 친화적 TUI 구축

### 작업 항목
- [ ] TUI 프레임워크 재설계
- [ ] CLI 명령 통합 레이어
- [ ] 인터랙티브 선택 UI
- [ ] 실시간 모니터링 뷰
- [ ] 작업 매니저 UI
- [ ] 설정 편집 UI

### TUI 아키텍처
```
TUI Layer (Bubble Tea)
    ↓
CLI Wrapper (명령 실행)
    ↓
CLI Commands (비즈니스 로직)
```

### 성공 지표
- TUI에서 모든 CLI 기능 사용 가능
- 부드러운 UI 전환
- 실시간 상태 업데이트
- 키보드 단축키 지원

## 일정 요약
- **Phase 1**: 1주 (1월 2주)
- **Phase 2**: 1주 (1월 3주)
- **Phase 3**: 2주 (1월 4주 - 2월 1주)
- **Phase 4**: 2주 (2월 2-3주)
- **총 기간**: 6주

## 장기 비전

### 추가 백업 타입
- PostgreSQL (pg_dump)
- MySQL (mysqldump)
- Elasticsearch (snapshot)
- Redis (RDB/AOF)

### 고급 기능
- 백업 스케줄링 (cron)
- 원격 스토리지 지원 (S3, GCS)
- 백업 검증 기능
- 증분 백업
- 웹 UI 버전

### 에코시스템
- Helm chart 제공
- Kubernetes Operator
- Prometheus 메트릭
- CI/CD 통합 예제

## 위험 관리

### 기술적 위험
- kubectl 의존성 → 추상화 레이어로 격리
- 대용량 백업 → 스트리밍 처리
- 네트워크 불안정 → 재시도 로직

### 사용자 경험
- CLI 복잡도 → 직관적 명령 설계
- 마이그레이션 → 명확한 가이드
- 하위 호환성 → 버전 관리

## 완화 전략
1. **점진적 개발**: 각 Phase별 동작 가능한 버전 배포
2. **사용자 피드백**: 조기 사용자 테스트
3. **문서화**: 명확한 사용 가이드
4. **호환성 레이어**: 기존 TUI 사용자를 위한 브릿지

## 마일스톤
- v0.2.0: CLI 핵심 완성
- v0.3.0: 아키텍처 정리
- v0.4.0: CLI 고도화
- v1.0.0: TUI 통합 완성