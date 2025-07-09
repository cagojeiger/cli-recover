# Phase 3.13: CLI 도구 자동 다운로드

## 개요
- **목표**: 외부 CLI 도구(kubectl, mc 등)가 없을 때 자동으로 다운로드
- **복잡도**: 50/100 (Occam's Razor 원칙 준수)
- **예상 기간**: 2025-01-11 (Phase 3.12 이후, Phase 4 이전)
- **상태**: 계획됨

## 배경
현재 cli-recover는 kubectl을 직접 호출하며, 향후 mc(MinIO Client)도 사용 예정입니다. 이러한 도구가 시스템에 없을 경우 자동으로 다운로드하여 사용자 경험을 개선합니다.

## 현재 문제점
1. **의존성 누락**: kubectl이 없으면 백업/복원 실패
2. **수동 설치 필요**: 사용자가 직접 도구 설치해야 함
3. **버전 불일치**: 시스템마다 다른 버전 사용 가능
4. **향후 확장성**: mc 등 새로운 도구 추가 시 같은 문제 반복

## 해결 방안

### 1. 도구 관리자 (ToolManager)
```go
// internal/infrastructure/tools/manager.go
type Manager struct {
    toolsDir string // ~/.cli-recover/tools/
}

// 주요 메서드
- EnsureKubectl(ctx) (string, error)
- EnsureMC(ctx) (string, error)
- Clean() error
```

### 2. 다운로드 전략
1. **우선순위 검색**:
   - 시스템 PATH 확인
   - ~/.cli-recover/tools/ 확인
   - 없으면 다운로드

2. **다운로드 URL 패턴**:
   - kubectl: `https://dl.k8s.io/release/{version}/bin/{os}/{arch}/kubectl`
   - mc: `https://dl.min.io/client/mc/release/{os}-{arch}/mc`

3. **플랫폼 감지**:
   ```go
   runtime.GOOS    // linux, darwin, windows
   runtime.GOARCH  // amd64, arm64
   ```

### 3. 안전한 다운로드
- 임시 파일 사용 (.tmp)
- 원자적 이동 (os.Rename)
- 실행 권한 설정 (Unix: 0755)
- HTTPS 전용

### 4. 통합 방법
```go
// kubernetes/kubectl.go 수정
func BuildKubectlCommand(args ...string) []string {
    kubectlPath, _ := toolManager.EnsureKubectl(context.Background())
    return append([]string{kubectlPath}, args...)
}
```

## 구현 상세

### Phase 1: 기본 구조 (복잡도: 20)
- ToolManager 생성
- 도구 검색 로직
- 디렉토리 관리

### Phase 2: kubectl 지원 (복잡도: +15 = 35)
- 최신 버전 조회 (stable.txt)
- 플랫폼별 다운로드
- 캐싱 메커니즘

### Phase 3: mc 지원 (복잡도: +10 = 45)
- MinIO 클라이언트 다운로드
- URL 패턴 처리

### Phase 4: 설정 통합 (복잡도: +5 = 50)
```yaml
tools:
  auto_download: true
  directory: ~/.cli-recover/tools
  kubectl:
    version: stable
  mc:
    version: latest
```

## TDD 계획

### 1. RED Phase
```go
// 도구가 없을 때 다운로드 테스트
TestEnsureKubectl_NotInPath_Downloads
TestEnsureKubectl_AlreadyDownloaded_ReusesExisting
TestDownloadFile_NetworkError_Cleanup
```

### 2. GREEN Phase
- Mock HTTP 클라이언트
- 최소 구현
- 에러 처리

### 3. REFACTOR Phase
- 진행률 표시 추가 고려
- 재시도 로직
- 로깅 개선

## 위험 요소 및 완화

### 1. 네트워크 문제
- **위험**: 다운로드 실패
- **완화**: 명확한 에러 메시지, 수동 설치 가이드 제공

### 2. 권한 문제
- **위험**: 도구 디렉토리 생성 실패
- **완화**: 사용자 홈 디렉토리 사용

### 3. 플랫폼 호환성
- **위험**: 지원하지 않는 OS/아키텍처
- **완화**: 주요 플랫폼만 지원 (linux/darwin/windows + amd64/arm64)

## 성공 지표
- [ ] kubectl 없는 환경에서 자동 다운로드
- [ ] 다운로드된 도구 재사용
- [ ] 에러 시 명확한 메시지
- [ ] 테스트 커버리지 90%
- [ ] 성능 영향 최소화

## 예상 사용 예시
```bash
# kubectl이 없는 환경
$ cli-recover backup filesystem nginx-pod /data
kubectl not found in PATH
Downloading kubectl v1.28.0... Done
Backup starting...

# 두 번째 실행 (캐시 사용)
$ cli-recover backup filesystem redis-pod /data
Backup starting...
```

## 대안 검토

### 1. 컨테이너 기반 접근 (복잡도: 80)
- Docker 이미지에 모든 도구 포함
- **단점**: Docker 의존성, 복잡도 증가

### 2. 정적 바이너리 포함 (복잡도: 90)
- 빌드 시 kubectl 포함
- **단점**: 바이너리 크기 증가, 라이선스 문제

### 3. 선택된 방안: 동적 다운로드 (복잡도: 50) ✅
- 필요시에만 다운로드
- 캐싱으로 재사용
- 단순하고 효과적

## 일정
- **계획**: 2025-01-08 (완료)
- **구현**: 2025-01-11 (Phase 3.12 이후)
- **예상 시간**: 4시간
  - ToolManager 구현: 2시간
  - 테스트 작성: 1.5시간
  - 통합 및 문서화: 0.5시간