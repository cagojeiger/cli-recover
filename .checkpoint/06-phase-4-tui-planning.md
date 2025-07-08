# Checkpoint: Phase 4 TUI 재구현 계획

## 날짜
- 계획 수립: 2025-01-08
- 예정 시작: 2025-01-09
- 목표 복잡도: 40/100

## 배경
- Phase 3.9에서 아키텍처 대폭 단순화 완료
- 3계층 → 2계층 구조로 변경
- Registry 패턴 제거, Factory 함수 사용
- 기존 TUI 코드는 이전 아키텍처 기반

## 재구현 필요성
1. **아키텍처 변경**
   - Application 레이어가 제거됨
   - 직접 Domain/Infrastructure 호출
   - 더 단순한 구조에 맞게 재작성

2. **단순화 원칙 적용**
   - 과도한 추상화 제거
   - 직접적인 호출 구조
   - Go 관용적 패턴 사용

## TUI 재구현 계획

### 1. 기본 구조
```
cmd/cli-recover/tui/
├── app.go          # TUI 메인 애플리케이션
├── menu.go         # 메인 메뉴
├── backup.go       # 백업 워크플로우
├── restore.go      # 복원 워크플로우  
├── list.go         # 백업 목록 표시
├── logs.go         # 로그 뷰어
└── progress.go     # 진행률 표시
```

### 2. 핵심 변경사항
- Executor 인터페이스 제거
- 직접 CLI 명령어 실행
- 단순한 이벤트 처리
- 최소한의 상태 관리

### 3. 구현 우선순위
1. **기본 메뉴 시스템**
   - tview 기본 설정
   - 메인 메뉴 네비게이션
   - 키보드 단축키

2. **백업 워크플로우**
   - Namespace 선택
   - Pod 선택
   - 경로 입력
   - 진행률 표시

3. **복원 워크플로우**
   - 백업 파일 선택
   - 대상 Pod 선택
   - 옵션 설정
   - 진행률 표시

4. **목록 및 로그**
   - 백업 목록 테이블
   - 로그 뷰어
   - 필터링 기능

## 설계 원칙

### 1. 단순성 우선
- 복잡한 상태 관리 피하기
- 직접적인 명령 실행
- 최소한의 추상화

### 2. CLI 래핑
```go
// Before: 복잡한 Executor 패턴
executor.Execute(ctx, job)

// After: 직접 실행
cmd := exec.Command("cli-recover", "backup", "filesystem", pod, path)
```

### 3. 에러 처리
- 사용자 친화적 메시지
- 복구 가능한 에러 처리
- 명확한 실패 원인 표시

## 예상 코드 구조

### app.go (메인 진입점)
```go
type App struct {
    *tview.Application
    pages  *tview.Pages
    config *config.Config
}
```

### backup.go (백업 워크플로우)
```go
type BackupView struct {
    form     *tview.Form
    progress *tview.TextView
    app      *App
}
```

## 성공 지표
- [ ] 모든 CLI 기능 TUI에서 사용 가능
- [ ] 부드러운 사용자 경험
- [ ] 복잡도 40/100 이하 유지
- [ ] 테스트 가능한 구조
- [ ] 명확한 에러 처리

## 위험 요소
1. **진행률 스트리밍**
   - CLI 출력 실시간 파싱
   - 별도 고루틴 관리

2. **상태 동기화**
   - TUI와 CLI 프로세스 간 동기화
   - 취소 처리

3. **크로스 플랫폼**
   - 터미널 호환성
   - 색상 및 스타일

## 참고사항
- tview 문서: https://github.com/rivo/tview
- 기존 TUI 코드는 참고용으로만 사용
- 단순화 원칙 항상 우선시
- 과도한 기능 추가 자제