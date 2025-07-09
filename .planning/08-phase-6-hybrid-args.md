# Phase 6: 하이브리드 인자 처리

## 개요
- **Phase**: 6
- **목표**: Positional Arguments와 Flags 동시 지원
- **복잡도**: 20/100
- **우선순위**: 낮음 (선택사항)
- **상태**: 계획

## 배경
Phase 3.12에서 분리된 기능으로, kubectl/docker 스타일의 유연한 인자 처리를 지원하기 위함.

## 목표
사용자가 동일한 정보를 positional arguments로도, flags로도 제공할 수 있게 하여 더 유연한 CLI 경험 제공.

## 설계

### 현재 방식 (Positional Only)
```bash
cli-recover backup filesystem nginx-pod /var/www
cli-recover restore filesystem nginx-pod backup.tar
```

### 하이브리드 방식 (목표)
```bash
# 모두 동일한 동작
cli-recover backup filesystem nginx-pod /var/www
cli-recover backup filesystem --pod=nginx-pod --path=/var/www
cli-recover backup filesystem nginx-pod --path=/var/www
cli-recover backup filesystem --pod=nginx-pod /var/www
```

## 구현 계획

### 1. 인자 파싱 로직 개선
```go
type ArgParser struct {
    positionals []string
    flags       map[string]string
}

func (p *ArgParser) GetPod() string {
    // 1. Flag 확인
    if pod := p.flags["pod"]; pod != "" {
        return pod
    }
    // 2. Positional 확인
    if len(p.positionals) > 0 {
        return p.positionals[0]
    }
    return ""
}
```

### 2. 우선순위 규칙
1. Flags가 있으면 우선 사용
2. Flags가 없으면 positional 사용
3. 둘 다 있으면 flag 우선
4. 충돌 시 경고 메시지

### 3. 명령어별 적용
- backup filesystem: pod, path
- restore filesystem: pod, backup-file
- list backups: namespace (이미 flag만 사용)

## 장단점

### 장점
- 더 유연한 사용성
- 스크립트에서 명확한 의미 전달
- 순서 기억할 필요 없음
- kubectl/docker와 일관성

### 단점
- 구현 복잡도 증가
- 중복 처리 로직 필요
- 문서화 복잡
- 테스트 케이스 증가

## 예상 작업량
- 파서 구현: 2-3시간
- 각 명령어 통합: 2-3시간
- 테스트 작성: 2-3시간
- 문서 업데이트: 1시간
- 총: 1-2일

## 결정 이유
**Phase 6로 연기한 이유:**
1. 현재 positional arguments만으로도 충분히 동작
2. CLAUDE.md Occam's Razor 원칙 - 불필요한 복잡도
3. 사용자 요구사항 없음
4. 더 중요한 작업들이 있음 (테스트 커버리지, TUI)

## 향후 검토 시점
- 사용자 피드백으로 요구사항 발생 시
- Phase 5 (테스트 커버리지) 완료 후
- CLI 사용성에 대한 불만 제기 시