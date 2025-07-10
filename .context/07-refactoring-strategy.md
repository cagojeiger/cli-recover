# Refactoring Strategy - 안전한 리팩토링 전략

## 문제 인식

과거 경험에서 같은 디렉토리 내 리팩토링으로 인한 혼란:
- 어떤 파일이 최신 버전인지 불명확
- 롤백 불가능한 상황 발생
- 팀원들의 혼란 가중
- Git blame이 의미를 잃음

## 핵심 원칙

### 1. Clear Boundaries (명확한 경계)
- 실험 코드는 `experimental/` 디렉토리에
- 프로덕션 코드는 절대 수정하지 않음
- 명확한 v2, v3 등의 버전 표시
- README.md로 실험 상태 명시

### 2. Parallel Development (병렬 개발)
- 기존 코드와 새 코드 공존
- 환경변수로 전환 제어
- A/B 테스트 가능한 구조
- 성능 비교 측정 가능

### 3. Gradual Migration (점진적 마이그레이션)
- 한 번에 하나의 기능만
- 성공 검증 후 다음 단계
- 실패 시 즉시 롤백 가능
- 모든 단계에서 프로덕션 안정성 유지

## 실행 전략

### Phase 1: Experimental Space
```
experimental/
└── providers/
    └── filesystem_v2/
        └── README.md  # "THIS IS EXPERIMENTAL" 명시
```

### Phase 2: Feature Toggle
```go
if os.Getenv("USE_EXPERIMENTAL") == "true" {
    log.Warn("Using EXPERIMENTAL provider")
    // 새로운 구현
} else {
    // 기존 구현 (변경 없음)
}
```

### Phase 3: Verification
- 충분한 테스트 기간
- 성능 비교
- 사용자 피드백
- 버그 추적

### Phase 4: Migration or Rollback
- 성공: 점진적 코드 이동
- 실패: experimental/ 삭제
- 부분 성공: 좋은 부분만 선택적 적용

## 실제 적용 예시

### 나쁜 예: 혼란스러운 리팩토링
```
internal/filesystem/
├── backup.go         # 이게 최신?
├── backup_old.go     # 아니면 이게 이전 버전?
├── backup_v2.go      # 설마 이건가?
└── backup_new.go     # 대체 뭐가 진짜?
```

### 좋은 예: 명확한 실험 구조
```
internal/filesystem/
└── backup.go         # 현재 프로덕션 (변경 없음)

experimental/
└── filesystem_v2/
    └── backup.go     # 실험 버전 (명확히 격리)
```

## 마이그레이션 체크리스트

### 실험 시작 전
- [ ] experimental/ 디렉토리 생성
- [ ] README.md에 실험 목적 명시
- [ ] 환경변수 설정 준비
- [ ] 롤백 계획 수립

### 실험 중
- [ ] 기존 코드 수정 금지 확인
- [ ] 테스트 커버리지 유지
- [ ] 성능 메트릭 수집
- [ ] 버그 및 이슈 추적

### 실험 완료 후
- [ ] 성공/실패 판단
- [ ] 이해관계자 합의
- [ ] 마이그레이션 또는 롤백 실행
- [ ] 문서화 및 교훈 기록

## 핵심 교훈

> "혼란스러운 깨끗한 코드보다 명확한 더러운 코드가 낫다"

리팩토링의 목적은 코드 개선이지만, 그 과정에서 팀의 생산성을 해치면 안 됩니다.
명확한 경계와 안전한 전환 전략이 성공적인 리팩토링의 핵심입니다.