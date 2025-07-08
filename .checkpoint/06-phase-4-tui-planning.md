# Checkpoint: Phase 4 TUI 구현 계획

## 날짜
- 계획: 2025-01-09~
- Phase 3.9 완료 후 시작

## 배경
- CLI 기능 완성됨
- 아키텍처 단순화 완료
- 사용자 친화적 인터페이스 필요

## 목표
- tview 기반 TUI 구현
- CLI 명령어 래핑 방식
- 실시간 진행률 표시
- 복잡도 40/100 이하 유지

## 구현 계획

### 1. TUI 구조
```
cmd/cli-recover/tui/
├── app.go          # TUI 메인 애플리케이션
├── menu.go         # 메인 메뉴 화면
├── backup/         # 백업 워크플로우
│   ├── select.go   # Provider/Pod/Path 선택
│   ├── options.go  # 백업 옵션 설정
│   └── progress.go # 진행률 표시
├── restore/        # 복원 워크플로우
│   ├── list.go     # 백업 목록 표시
│   ├── select.go   # 대상 선택
│   └── progress.go # 진행률 표시
├── logs/           # 로그 뷰어
│   ├── list.go     # 로그 목록
│   └── viewer.go   # 로그 상세 보기
└── common/         # 공통 컴포넌트
    ├── styles.go   # UI 스타일
    ├── widgets.go  # 재사용 위젯
    └── helpers.go  # 유틸리티
```

### 2. 화면 흐름
```
Main Menu
├─ Backup
│  ├─ Select Type (filesystem)
│  ├─ Select Namespace
│  ├─ Select Pod
│  ├─ Enter Path
│  ├─ Set Options
│  └─ Execute with Progress
├─ Restore
│  ├─ List Backups
│  ├─ Select Backup
│  ├─ Select Target Pod
│  ├─ Confirm
│  └─ Execute with Progress
├─ List Backups
│  └─ Browse with Details
├─ View Logs
│  ├─ List Logs
│  └─ Show Log Details
└─ Settings
   └─ Configuration
```

### 3. 기술적 접근

#### CLI 래핑
```go
// TUI에서 CLI 명령 실행
func executeBackup(opts BackupOptions) error {
    cmd := exec.Command("cli-recover", "backup", "filesystem", 
        opts.Pod, opts.Path, "--namespace", opts.Namespace)
    
    // 진행률 파싱
    stdout, _ := cmd.StdoutPipe()
    go parseProgress(stdout, progressChan)
    
    return cmd.Run()
}
```

#### 진행률 표시
- CLI의 stderr 출력 파싱
- [PROGRESS] 태그 감지
- 실시간 UI 업데이트

### 4. UI 컴포넌트

#### 메인 메뉴
- List 위젯 사용
- 키보드 단축키 지원
- 상태바에 도움말

#### 진행률 화면
- ProgressBar 위젯
- 로그 출력 표시
- 취소 가능

#### 목록 브라우저
- Table 위젯 사용
- 정렬/필터 기능
- 상세 정보 패널

### 5. 예상 코드량
- app.go: ~100줄
- menu.go: ~80줄
- 각 워크플로우: ~150줄
- 공통 컴포넌트: ~200줄
- **총합**: ~800줄 목표

## 위험 관리
1. **복잡도 관리**: 단순한 래핑 유지
2. **에러 처리**: CLI 출력 파싱 주의
3. **성능**: 대용량 목록 처리

## 성공 지표
- [ ] 모든 CLI 기능 TUI에서 사용 가능
- [ ] 부드러운 UI 전환
- [ ] 실시간 진행률 표시
- [ ] 키보드 단축키 완성
- [ ] 복잡도 40/100 이하

## 참고사항
- Phase 3.9로 코드베이스가 단순해짐
- Factory 패턴으로 Provider 생성 용이
- 기존 CLI 로직 재사용 가능