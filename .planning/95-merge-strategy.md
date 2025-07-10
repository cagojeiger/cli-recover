# 기존 컨텍스트와 병합 전략

## 현재 상태 분석

### 기존 문서 구조
```
.planning/
├── 00-roadmap.md                    # 전체 로드맵
├── 01-current-sprint.md             # 현재 스프린트
├── 02-backlog.md                    # 백로그
├── 03-1-restore-fix.md              # Phase 3-1 긴급 수정
├── 04-phase-3.12-cli-usability.md   # CLI 사용성 개선
├── 06-phase-3.13-tool-auto-download.md # 도구 자동 다운로드
├── 08-phase-6-hybrid-args.md        # 하이브리드 인자 처리
└── README.md                         # 설명서
```

### 새로 작성한 문서
```
.planning/
├── 99-vision-draft.md               # 새 비전 (임시)
├── 98-architecture-insights.md      # 아키텍처 통찰
├── 97-provider-analysis.md          # Provider 분석
├── 96-provider-patterns-research.md # 사용 패턴 연구
└── 95-merge-strategy.md             # 병합 전략 (현재)

.memory/short-term/
└── 05-architecture-pivot.md         # 전환점 기록
```

## 병합 계획

### Phase 1: 즉시 병합 가능한 것들

#### 1. `00-roadmap.md` 업데이트
**변경 내용**:
- Phase 4를 "TUI 구현"에서 "아키텍처 리팩토링"으로 변경
- Phase 5를 "테스트 커버리지"에서 "CLI/TUI 통합"으로 변경
- Phase 6를 "하이브리드 인자"에서 "Provider 확장"으로 변경
- 기존 Phase들은 7, 8, 9로 밀림

**이유**: 아키텍처 개선이 선행되어야 나머지가 의미 있음

#### 2. `01-current-sprint.md` 수정
**추가 항목**:
- [ ] Provider 독립 구조 설계
- [ ] CLI Definition 표준 정의
- [ ] Filesystem Provider 리팩토링
- [ ] 기존 코드 마이그레이션 계획

#### 3. `02-backlog.md` 재정렬
**우선순위 변경**:
- Provider 독립 아키텍처 (최우선)
- CLI/TUI 통합 시스템
- MongoDB Provider
- MinIO Provider
- (기존 항목들은 하위로)

### Phase 2: 새로운 컨텍스트 구조

#### `.context/` 디렉토리 생성
```
.context/
├── 00-project.md          # 프로젝트 개요 (기존 + 새 비전)
├── 01-architecture.md     # 아키텍처 (Provider 독립 구조)
├── 02-tech-stack.md       # 기술 스택 (변경 없음)
├── 03-patterns.md         # 코딩 패턴 (Provider별 특화)
└── 04-decisions.md        # 아키텍처 결정 기록
```

#### `.memory/long-term/` 업데이트
```
.memory/long-term/
├── 00-decisions.md        # 주요 결정 사항
├── 01-learnings.md        # 배운 점들
├── 02-patterns.md         # 발견한 패턴들
└── 03-architecture-evolution.md  # 아키텍처 진화 과정
```

### Phase 3: 점진적 마이그레이션

#### 1단계: 구조 전환 (1주)
- [ ] Provider 인터페이스 제거
- [ ] Filesystem Provider 독립 구현
- [ ] 기존 테스트 유지하며 리팩토링

#### 2단계: CLI 재설계 (1주)
- [ ] CLI Definition 구현
- [ ] 명령어 구조 변경 (fs backup, mongo dump 등)
- [ ] 하위 호환성 레이어

#### 3단계: TUI 통합 (2주)
- [ ] 자동 TUI 생성 시스템
- [ ] 기존 TUI 코드 마이그레이션
- [ ] 통합 테스트

## 충돌 해결 전략

### 1. 기존 Phase와의 충돌
- **문제**: Phase 4가 TUI vs 아키텍처
- **해결**: TUI는 아키텍처 개선 후가 더 효율적
- **조치**: Phase 순서 재배열

### 2. 인터페이스 변경
- **문제**: 기존 Provider 인터페이스 사용 코드
- **해결**: Adapter 패턴으로 임시 호환
- **조치**: 점진적 마이그레이션

### 3. 명령어 구조 변경
- **문제**: backup filesystem → fs backup
- **해결**: 둘 다 지원하다가 v2에서 제거
- **조치**: Deprecation 경고

## 위험 관리

### 1. 기존 사용자 영향
- v1 명령어 계속 지원
- 마이그레이션 가이드 제공
- 충분한 전환 기간

### 2. 개발 일정 지연
- 핵심 기능만 먼저 전환
- Provider별 순차 적용
- 기존 코드 유지하며 병행

### 3. 테스트 커버리지 하락
- 리팩토링 전 테스트 작성
- Provider별 독립 테스트
- 통합 테스트는 최소화

## 실행 순서

### 즉시 (오늘)
1. ✅ 임시 문서들 작성 완료
2. ⏳ 00-roadmap.md 업데이트
3. ⏳ 01-current-sprint.md 수정

### 단기 (이번 주)
1. .context/ 디렉토리 구조 생성
2. 아키텍처 결정 문서 작성
3. Filesystem Provider 프로토타입

### 중기 (2-4주)
1. Provider 독립 구조 구현
2. CLI Definition 시스템
3. TUI 자동 생성

### 장기 (1-2개월)
1. 모든 Provider 마이그레이션
2. v2.0 릴리스
3. v1 deprecation

## 성공 지표

### 기술적
- [ ] Provider별 독립 테스트 가능
- [ ] CLI/TUI 코드 중복 제거
- [ ] 새 Provider 추가 시간 < 1일

### 사용자 경험
- [ ] 기존 명령어 계속 동작
- [ ] 더 직관적인 명령어 구조
- [ ] Provider별 특화 기능 활용

### 유지보수
- [ ] 코드 복잡도 감소
- [ ] 문서화 개선
- [ ] 기여자 진입 장벽 낮춤

## 다음 단계

1. **즉시**: 이 병합 전략 검토 및 승인
2. **오늘**: 00-roadmap.md 업데이트 시작
3. **내일**: .context/ 구조 생성 및 문서 작성
4. **이번 주**: Filesystem Provider 프로토타입

이 전략을 통해 기존 작업을 보존하면서도
새로운 비전을 향해 나아갈 수 있습니다.