# Lessons Learned from v1 Implementation

## 배경
v1 구현 과정에서 많은 실험과 리팩토링을 거쳤습니다. 이 문서는 그 과정에서 얻은 교훈을 정리하여 v2 구현에 적용하기 위한 것입니다.

## 핵심 철학: Isolation > Reusability

### 발견된 문제
- 공통 인터페이스를 강제하니 각 Provider의 특성을 살리기 어려웠음
- 재사용을 위한 추상화가 오히려 복잡성을 증가시킴
- Provider 간 의존성이 생겨 수정이 어려워짐

### 해결책
> "Duplication is cheaper than the wrong abstraction" — Sandi Metz

- 각 Provider는 완전히 독립적으로 구현
- 코드 중복이 있더라도 격리를 우선시
- 필요한 경우 복사하되 공유하지 않음

## 아키텍처 교훈

### 1. 과도한 계층화의 함정
- **문제**: 3계층 아키텍처 (Domain/Application/Infrastructure)가 과도함
- **결과**: Application 레이어가 단순 pass-through 역할만 함
- **해결**: 2계층으로 단순화하여 복잡도 45% 감소

### 2. God Object 회피
- **문제**: TUI Model이 115개 이상의 필드를 가진 거대 객체가 됨
- **원인**: 모든 상태를 하나의 구조체에 담으려 함
- **교훈**: 작은 단위로 분리하고 필요한 것만 전달

### 3. Factory > Registry
- **문제**: Provider Registry 패턴이 불필요하게 복잡
- **해결**: 단순한 Factory 함수로 충분
- **이유**: Provider 개수가 적고 동적 등록이 불필요

## 기술적 교훈

### 1. 원자적 파일 쓰기
```go
// 잘못된 방법
file, _ := os.Create(outputPath)
// 백업 중단 시 불완전한 파일 남음

// 올바른 방법
tmpFile := outputPath + ".tmp"
file, _ := os.Create(tmpFile)
// 작업 완료 후
os.Rename(tmpFile, outputPath)
```

### 2. 스트리밍 처리의 중요성
- 대용량 파일도 메모리 효율적 처리 필요
- `io.TeeReader`로 진행률 보고와 처리 동시 수행
- 버퍼 크기 조정으로 성능 최적화

### 3. 표준 라이브러리 우선
- 외부 의존성 최소화로 안정성 향상
- 대부분의 기능은 Go 표준 라이브러리로 충분
- 필요시 작은 헬퍼 함수로 해결

## 개발 프로세스 교훈

### 1. TDD의 실제 효과
- 안전한 리팩토링 가능
- 문서화 효과 (테스트가 곧 사용 예제)
- 설계 개선 유도 (테스트하기 쉬운 코드가 좋은 코드)

### 2. 구조적 변경과 행동 변경 분리
```bash
# 좋은 커밋 히스토리
git log --oneline
abc1234 test: add test for backup progress
def5678 feat: implement backup progress
ghi9012 refactor: extract progress reporter
```

### 3. 계획의 중요성
- 구현 전 설계 문서 작성
- 복잡도 예측 및 관리
- 명확한 성공 기준 설정

## 리팩토링 트라우마

### 문제 상황
```
filesystem/
├── backup.go       # 현재 버전?
├── backup_old.go   # 이전 버전?
├── backup_v2.go    # 새 버전?
└── backup_test.go  # 어떤 버전 테스트?
```

### 교훈
- 같은 디렉토리에서 리팩토링 시 혼란 발생
- Git history도 의미를 잃음
- 어떤 코드가 production인지 불분명

### v2에서의 접근
- experimental/ 디렉토리로 명확한 경계
- 기존 코드는 절대 수정하지 않음
- 성공 검증 후에만 migration
- 환경변수로 안전한 전환

## CLI vs TUI

### CLI-First의 장점
- 자동화 가능
- 테스트 용이
- 디버깅 간편
- Unix 철학 준수

### TUI는 CLI의 Wrapper
- TUI가 CLI 명령을 내부적으로 호출
- 일관된 동작 보장
- 유지보수 부담 감소

## 진행률 보고의 중요성

### 3초 규칙
- 3초 이상 걸리는 작업은 반드시 진행률 표시
- 사용자 불안감 해소
- 작업 취소 가능성 제공

### 통합된 시스템
- Terminal, CI/CD, Pipe 환경 모두 지원
- 일관된 형식과 메시지
- 표준 출력 활용

## v2 구현 가이드라인

### DO
- ✅ 완전히 격리된 구현
- ✅ TDD로 시작
- ✅ 단순한 것부터 구현
- ✅ 명확한 에러 메시지
- ✅ 진행률 보고 내장

### DON'T
- ❌ 조기 추상화
- ❌ 공유 인터페이스
- ❌ 복잡한 의존성
- ❌ 과도한 설정
- ❌ 불명확한 경계

## 성공의 정의
- 코드를 읽는 사람이 즉시 이해 가능
- 수정이 두렵지 않은 구조
- 테스트가 신뢰할 수 있는 안전망
- 사용자가 만족하는 도구