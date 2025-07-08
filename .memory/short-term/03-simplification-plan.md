# Phase 3.9: 단순화 실행 계획

## 배경
- 현재 복잡도: 75/100 (과도함)
- 목표 복잡도: 30/100
- 핵심 원칙: Occam's Razor

## 문제 분석

### 1. 과도한 레이어링
```
현재: cmd → application/adapters → domain → infrastructure
목표: cmd → domain/infrastructure
```

### 2. 중복된 도메인
- domain/backup과 domain/restore가 거의 동일
- 각각 provider, registry, types 중복

### 3. 미사용 코드
- minio/mongodb 스텁만 존재
- runner 패키지 용도 불명확
- 중복 테스트 파일

### 4. 불필요한 추상화
- Registry 패턴 (provider 1개만 사용)
- 과도한 인터페이스 분리

## 실행 계획

### Step 1: Application 레이어 제거
1. **Config 이동**
   ```bash
   mv internal/application/config internal/infrastructure/
   ```

2. **Adapter 통합**
   - backup_adapter.go → cmd/cli-recover/backup.go
   - restore_adapter.go → cmd/cli-recover/restore.go
   - list_adapter.go → cmd/cli-recover/list.go

3. **Application 디렉토리 삭제**
   ```bash
   rm -rf internal/application
   ```

### Step 2: Domain 통합
1. **Operation 도메인 생성**
   ```bash
   mkdir internal/domain/operation
   ```

2. **Backup/Restore 통합**
   - 공통 Provider interface
   - 통합된 Types
   - Registry 제거

3. **기존 도메인 삭제**
   ```bash
   rm -rf internal/domain/backup
   rm -rf internal/domain/restore
   ```

### Step 3: 미사용 코드 제거
1. **Runner 패키지**
   ```bash
   rm -rf internal/infrastructure/runner
   ```

2. **MinIO/MongoDB 스텁**
   - backup.go의 case 문 제거
   - restore.go의 case 문 제거

3. **중복 테스트**
   - restore_adapter_coverage_test.go

### Step 4: 구조 평탄화
1. **Providers 디렉토리 제거**
   ```bash
   mv internal/infrastructure/providers/filesystem internal/infrastructure/
   rm -rf internal/infrastructure/providers
   ```

2. **Import 경로 수정**
   - 모든 파일의 import 경로 업데이트

## 예상 결과

### 디렉토리 구조
```
internal/
├── domain/
│   ├── operation/   # 통합
│   ├── log/
│   ├── logger/
│   └── metadata/
└── infrastructure/
    ├── config/      # 이동
    ├── filesystem/  # 평탄화
    ├── kubernetes/
    └── logger/
```

### 메트릭스
- 파일 수: ~61개 → ~37개 (-40%)
- 디렉토리: 17개 → 10개 (-41%)
- 코드 라인: 예상 -35%
- 복잡도: 75 → 30

## 위험 관리
1. **각 단계별 테스트**
   ```bash
   go test ./...
   ```

2. **Git 커밋 전략**
   - 각 Step별 개별 커밋
   - 롤백 가능한 단위

3. **기능 검증**
   - CLI 명령어 동작 확인
   - TUI 기능 테스트

## 성공 기준
- ✅ 모든 테스트 통과
- ✅ 기능 변경 없음
- ✅ 코드 가독성 향상
- ✅ 복잡도 30 이하

## 최종 완료 (2025-01-08)
모든 단계가 성공적으로 완료되었습니다:
- Step 1: Application 레이어 제거 ✅
- Step 2: Domain 통합 ✅ (Registry 패턴 제거)
- Step 3: 미사용 코드 제거 ✅
- Step 4: 구조 평탄화 ✅

### 달성된 결과
- 파일 수: ~40% 감소
- 디렉토리 구조: 3단계로 평탄화
- 복잡도: 75 → ~30 (목표 달성!)
- 모든 테스트 통과
- 빌드 성공
- backup 디렉토리 제거
