# Architecture Pivot Point - Phase 3 완료 후 전환점

## 전환점 도달
- 시점: Phase 3-1 urgent fix 완료 후
- 계기: restore 명령어가 "File exists" 에러로 실패
- 통찰: filesystem restore는 본질적으로 `cp`에 가까움

## 핵심 깨달음

### 1. "One Interface to Rule Them All"의 함정
- 시도: 모든 Provider가 같은 인터페이스 구현
- 결과: 각 Provider의 특성을 살리지 못함
- 예시: MongoDB의 oplog vs Filesystem의 tar

### 2. Restore ≠ Backup⁻¹
- 가정: restore는 backup의 역연산
- 현실: 
  - Backup: 아카이브 생성
  - Restore: 선택적 파일 복사
- 사용자는 전체가 아닌 일부만 복원하길 원함

### 3. CLI/TUI 분리의 비효율
- 현재: CLI와 TUI가 별도 구현
- 문제: 동일한 기능을 두 번 구현
- 해법: 메타데이터 기반 자동 생성

## 결정 사항

### 1. Provider 독립 아키텍처로 전환
- 이유: 각 Provider의 특성 최대한 활용
- 방법: 공통 인터페이스 제거, Provider별 최적화

### 2. CLI 명령어 재설계
- 기존: `backup <type>` 구조
- 신규: `<type> <action>` 구조
- 예: `fs backup`, `mongo dump`, `s3 sync`

### 3. 점진적 마이그레이션
- Phase 4: 구조 전환
- Phase 5: CLI/TUI 통합
- Phase 6: 새 Provider 추가

## 위험 요소

### 1. Breaking Changes
- v1 사용자 호환성
- 해결: 호환 레이어 제공

### 2. 코드 중복
- Provider별 독립 = 일부 중복
- 판단: 중복보다 특화가 더 중요

### 3. 복잡도 증가
- 여러 Provider 관리
- 해결: 명확한 구조와 문서화

## 교훈

### 1. 조기 추상화의 위험
- "나중에 필요할 것 같아서" 만든 추상화
- 실제로는 제약이 됨

### 2. 사용자 관점의 중요성
- 개발자: "깔끔한 인터페이스"
- 사용자: "내가 아는 명령어 스타일"

### 3. 도메인 지식의 가치
- tar, mongodump, s3 각각의 특성 이해
- 일반화보다 특화가 더 가치 있음

## 다음 단계

### 즉시
- 임시 비전 문서 작성 완료
- Provider별 실제 사용 패턴 조사

### 단기
- 기존 컨텍스트와 병합 계획
- Phase 4 상세 설계

### 장기
- Provider 플러그인 시스템
- 선언적 백업 정책

## 메모
- 2025-01-10: 아키텍처 전환 결정
- restore hanging 이슈가 큰 그림을 보게 함
- "완벽한 추상화"보다 "실용적 구현" 선택