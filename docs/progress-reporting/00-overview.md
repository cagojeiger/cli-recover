# 진행률 보고 시스템 개요

## 목적

CLI 도구에서 장시간 실행되는 작업의 진행 상황을 사용자에게 효과적으로 전달하는 것은 필수적입니다. 특히 백업, 복원, 다운로드와 같은 작업에서는 더욱 중요합니다.

## 핵심 원칙

### 1. 3초 규칙
**3초 이상 걸리는 모든 작업은 진행률을 보고해야 합니다.**

이유:
- 사용자는 3초 이상 아무 반응이 없으면 프로그램이 멈췄다고 생각
- 작업이 진행 중임을 시각적으로 확인할 수 있어야 함
- 예상 완료 시간을 알 수 있으면 더 좋음

### 2. 다중 환경 지원
하나의 구현으로 모든 환경을 지원해야 합니다:

```
터미널 (대화형) → 실시간 업데이트 with \r
CI/CD 파이프라인 → 주기적 로그 출력
로그 파일 → 구조화된 진행률 기록
TUI 인터페이스 → 별도 채널로 데이터 전송
```

### 3. 복잡도 최소화
- 외부 라이브러리 의존성 없음
- Go 표준 라이브러리만으로 구현
- 복잡도 증가: +5/100 (Occam's Razor 준수)

## 진행률 표시 유형

### 1. 확정적 진행률 (Determinate Progress)
전체 크기를 알 때:
```
Downloading kubectl... [████████░░░░] 67% (33.5MB/50MB) ETA: 15s
```

### 2. 불확정 진행률 (Indeterminate Progress)
전체 크기를 모를 때:
```
Processing... 125MB processed (5.2MB/s)
```

### 3. 단계별 진행률 (Stepped Progress)
여러 단계가 있을 때:
```
[1/4] Connecting to cluster... Done
[2/4] Estimating size... Done
[3/4] Creating backup... In Progress
[4/4] Verifying checksum...
```

## 구현 패턴

### 기본 인터페이스
```go
type ProgressReporter interface {
    // 진행률 업데이트
    Update(current, total int64, message string)
    
    // 작업 완료
    Complete(message string)
    
    // 에러 발생
    Error(err error)
}
```

### 환경 감지
```go
import "golang.org/x/term"

func isInteractive() bool {
    return term.IsTerminal(int(os.Stderr.Fd()))
}
```

## 적용 대상

### 필수 적용
- 파일시스템 백업/복원
- 네트워크 다운로드
- 대용량 데이터 처리
- 외부 명령 실행 (kubectl exec 등)

### 선택 적용
- 빠른 API 호출 (< 3초)
- 작은 파일 작업
- 메타데이터 조회

## 사용자 경험 향상

### 좋은 진행률 표시
```
✅ 남은 시간 예측 (ETA)
✅ 전송 속도 표시
✅ 완료 비율
✅ 실제 데이터 크기
```

### 피해야 할 패턴
```
❌ 변화 없는 메시지 ("Processing...")
❌ 너무 빠른 업데이트 (100ms 이하)
❌ 부정확한 예측
❌ 터미널 깨짐 문자
```

## 성능 고려사항

### 업데이트 빈도
- 터미널: 100ms ~ 500ms 간격
- 로그: 10초 간격
- TUI: 실시간 (throttled)

### 메모리 사용
- 진행률 추적을 위한 오버헤드 최소화
- 버퍼링 없이 스트리밍 처리
- io.TeeReader 패턴 활용

## 예시 시나리오

### 백업 작업
```
$ cli-recover backup filesystem nginx-pod /data
Estimating backup size... Done (1.2GB)
Creating backup... [██████████░░░░░░] 67% (804MB/1.2GB) ETA: 45s
```

### 도구 다운로드
```
$ cli-recover backup filesystem nginx-pod /data
kubectl not found, downloading...
Downloading kubectl v1.28.0... [████████████████] 100% (50MB/50MB)
Installation complete!
```

### CI/CD 환경
```
[2025-01-08 10:30:00] Starting backup for nginx-pod
[2025-01-08 10:30:10] Progress: 10% complete (120MB/1.2GB)
[2025-01-08 10:30:20] Progress: 20% complete (240MB/1.2GB)
...
[2025-01-08 10:31:40] Backup completed successfully
```

## 다음 단계

1. [구현 가이드](01-implementation-guide.md) - 실제 코드 작성 방법
2. [예제 모음](02-examples.md) - 다양한 사용 사례와 출력 예시