# 아키텍처 정리 작업 (2025-01-09)

## 수행한 작업

### 1. 플래그 충돌 해결
- restore_test.go 수정: overwrite → force

### 2. storage 패키지 이동
- `domain/log/storage/` → `infrastructure/log/storage/`
- 도메인에서 infrastructure 개념 제거

### 3. metadata 패키지 분리
- **Domain**: Store 인터페이스만 (`store_interface.go`)
- **Infrastructure**: FileStore 구현체 (`file_store.go`)
- 테스트 파일들도 infrastructure로 이동

### 4. operation 패키지 확인
- 실제로 사용되지 않음 (자체 테스트에서만 사용)
- 제거 대상으로 분류 (낮은 우선순위)

## 아키텍처 개선 결과

### 의존성 방향 (✅ 올바름)
```
CMD 
 ↓
Infrastructure (구현체)
 ↓  
Domain (인터페이스)
```

### Domain 순수성
- 외부 의존성 없음
- 인터페이스와 타입만 정의
- 비즈니스 로직만 포함

### Infrastructure 책임
- Domain 인터페이스 구현
- 외부 시스템 연동
- 파일/네트워크 I/O
- 실제 저장소 구현

## 복잡도 변화
- 이전: 30/100
- 현재: 25/100 (더 단순해짐)

## 다음 고려사항
- operation 패키지 제거 검토
- 테스트 커버리지는 현재 수준(50.7%) 유지
- 오캄의 면도날 원칙 준수