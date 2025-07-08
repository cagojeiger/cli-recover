# Phase 3.10: 백업 파일 무결성 보장

## 개요
- **목표**: 백업 중 파일 손상 방지를 위한 원자적 파일 쓰기 구현
- **복잡도**: 35/100 (체크섬 포함) ✅
- **원칙**: Occam's Razor - 가장 단순하면서 효과적인 해결책
- **예상 기간**: 2025-01-09

## 현재 문제점 분석

### 1. 불완전한 파일 쓰기 (가장 크리티컬)
**위험도**: 95/100
```go
// filesystem.go line 113
written, err := io.Copy(outputFile, stdout)
```
- **문제**: 프로세스 중단 시 불완전한 tar 파일 생성
- **시나리오**: 
  - 10GB 백업 중 7GB에서 네트워크 끊김
  - `backup-2024-01-08.tar` 파일에 7GB만 존재
  - 사용자는 완전한지 불완전한지 알 수 없음

### 2. 무결성 검증 타이밍
**위험도**: 80/100
```go
// backup_logic.go line 413-414
checksum, err := calculateFileChecksum(opts.OutputFile)  // 백업 완료 후에만 계산
```
- **문제**: 손상된 파일을 정상으로 착각
- **시나리오**: 중간에 실패했지만 체크섬이 계산되지 않음

### 3. 에러 시 파일 잔존
**위험도**: 70/100
```go
// filesystem.go line 88-92
outputFile, err := os.Create(opts.OutputFile)
defer outputFile.Close()  // 에러 발생해도 파일은 남음
```
- **문제**: 실패한 백업을 성공으로 착각 가능

## 해결책: 임시 파일 + 원자적 이동

### 왜 이 방법이 효과적인가?

#### 1. 임시 파일 사용
```go
// 현재 (위험)
outputFile, err := os.Create("backup.tar")
io.Copy(outputFile, stdout)  // 중간에 실패하면?

// 개선 (안전)
tempFile := "backup.tar.tmp"
outputFile, err := os.Create(tempFile)
io.Copy(outputFile, stdout)  // 실패해도 .tmp만 손상
```

#### 2. 원자적 이동 (Atomic Rename)
```go
// OS 레벨에서 원자적으로 처리
os.Rename("backup.tar.tmp", "backup.tar")
```
- **원자성**: "모두 성공하거나 모두 실패" - 중간 상태 없음
- **성능**: 같은 파일시스템 내에서는 메타데이터만 변경 (매우 빠름)
- **안전성**: 실패 시 원본 파일 영향 없음

### 실제 효과

**케이스 1: 프로세스 중단**
```bash
# 현재
$ cli-recover backup filesystem pod /data
^C  # Ctrl+C로 중단
$ ls -la
backup.tar  # 5GB/10GB 손상된 파일

# 개선 후
$ cli-recover backup filesystem pod /data
^C  # Ctrl+C로 중단
$ ls -la
backup.tar.tmp  # 임시 파일만 존재, 최종 파일 없음
```

**케이스 2: 디스크 공간 부족**
```bash
# 현재
Error: no space left on device
$ tar -tf backup.tar
tar: Unexpected EOF  # 손상된 아카이브

# 개선 후
Error: no space left on device
$ ls backup.tar
ls: cannot access 'backup.tar': No such file
```

## 구현 계획

### 1. filesystem.go 수정
```go
func (p *Provider) Execute(ctx context.Context, opts backup.Options) error {
    // ... 기존 코드 ...
    
    // 임시 파일 경로
    tempFile := opts.OutputFile + ".tmp"
    
    // defer로 실패 시 정리
    var success bool
    defer func() {
        if !success {
            os.Remove(tempFile)
        }
    }()
    
    // 임시 파일에 쓰기
    outputFile, err := os.Create(tempFile)
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    defer outputFile.Close()
    
    // ... io.Copy 등 기존 로직 ...
    
    // 성공 시 원자적 이동
    if err := os.Rename(tempFile, opts.OutputFile); err != nil {
        return fmt.Errorf("failed to finalize backup: %w", err)
    }
    
    success = true
    return nil
}
```

### 2. 테스트 추가
- 중간 중단 시나리오 테스트
- 디스크 공간 부족 테스트
- 동시 백업 실행 테스트
- 원자적 이동 실패 케이스

### 3. 복원 시 검증 강화
- .tmp 파일 무시
- 완전한 tar 파일만 표시
- 체크섬 검증 필수화

## 복잡도 평가

### 현재 해결책: 25/100 ✅
- 코드 추가: 3-4줄
- 외부 의존성: 없음
- 기존 로직 변경: 최소
- 테스트 가능성: 높음

### 대안 비교
1. **파일 잠금 (복잡도 50)**
   - OS별 구현 차이
   - 추가 라이브러리 필요
   
2. **체크섬 스트리밍 (복잡도 60)**
   - TeeReader 구현
   - 성능 오버헤드
   
3. **트랜잭션 로그 (복잡도 80)**
   - 별도 상태 관리
   - 복구 로직 복잡

## 구현 상세 (TDD 접근)

### Phase 1: 파일 손상 방지 (복잡도: 25)
1. **RED**: 백업 중단 시 최종 파일 없음 테스트
2. **GREEN**: 임시 파일 + os.Rename 구현
3. **REFACTOR**: defer 정리 로직 추가

### Phase 2: 데이터 무결성 (복잡도: +10 = 35)
1. **RED**: 체크섬 불일치 감지 테스트
2. **GREEN**: ChecksumWriter 구현
3. **REFACTOR**: 스트리밍 체크섬 통합

### Phase 3: 백업 검증 (선택사항, 복잡도: +10 = 45)
1. **RED**: tar 무결성 검증 테스트
2. **GREEN**: 파일 수, 크기 추적
3. **REFACTOR**: 메타데이터 시스템 통합

## TDD 테스트 시나리오

### 필수 테스트 케이스
```go
// 1. 백업 중단 시 임시 파일만 존재
TestBackupInterruption_LeavesOnlyTempFile

// 2. 성공 시 최종 파일만 존재
TestBackupSuccess_OnlyFinalFileExists

// 3. 체크섬 정확성
TestChecksum_CalculatedDuringStreaming

// 4. 디스크 공간 부족
TestDiskSpaceFull_CleansUpTempFile

// 5. 동시 백업 방지
TestConcurrentBackups_PreventedForSameFile
```

### 통합 테스트
```go
// 실제 파일시스템 테스트
TestIntegration_RealFileSystem

// 대용량 파일 테스트
TestLargeFile_ConstantMemoryUsage
```

## 성공 지표
- [ ] 중단된 백업이 최종 파일을 생성하지 않음
- [ ] 실패한 백업의 임시 파일 자동 정리
- [ ] 체크섬이 스트리밍 중 계산됨
- [ ] 메모리 사용량이 파일 크기와 무관
- [ ] 모든 기존 테스트 통과
- [ ] 테스트 커버리지 90% 이상
- [ ] 성능 영향 5% 이하

## 위험 요소
- 파일시스템 간 이동 시 rename 실패 가능
  - 해결: 같은 디렉토리에 임시 파일 생성
- 임시 파일 이름 충돌
  - 해결: PID 또는 타임스탬프 추가 고려

## 일정
- 구현: 1시간
- 테스트 작성: 2시간
- 통합 테스트: 1시간
- 문서 업데이트: 30분
- **총 예상 시간**: 4.5시간