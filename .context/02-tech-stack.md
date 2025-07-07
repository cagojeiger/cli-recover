# 기술 스택

## 프로그래밍 언어
- **Go 1.24.3**: 메인 언어
  - 크로스 플랫폼 지원
  - 정적 타입 안정성
  - 뛰어난 동시성 지원

## CLI 프레임워크
- **Cobra**: 명령어 라인 파싱
  - 서브커맨드 지원
  - 플래그 관리
  - 자동 help 생성
- **YAML**: 설정 파일 (gopkg.in/yaml.v3)
  - 설정 파일 파싱
  - 구조화된 설정 관리

## 테스트 프레임워크
- **testify**: 어설션 라이브러리
  - assert/require 패키지
  - mock 생성 도구
- **golden files**: 스냅샷 테스트
  - kubectl 출력 모킹
  - 재현 가능한 테스트

## 로깅 시스템
- **구조화된 로거**: 자체 구현
  - 파일/콘솔 출력
  - 로그 로테이션
  - JSON/Text 포맷
  - 레벨 필터링

## 백업 도구
- **tar**: 파일시스템 백업
  - 압축 옵션 (gzip, bzip2, xz)
  - 제외 패턴 지원
  - 진행률 추적

## 외부 도구
- **kubectl**: Kubernetes CLI
  - 필수 의존성
  - exec/cp 명령 사용
- **tar**: 아카이브 도구
  - 모든 Unix 시스템 표준

## 개발 도구
- **Make**: 빌드 자동화
  - 테스트, 빌드, 설치
  - 크로스 컴파일
- **golangci-lint**: 코드 품질
  - 정적 분석
  - 포맷팅 검사

## 아키텍처 패턴
- **Hexagonal Architecture**: 핵심 아키텍처
  - Domain/Infrastructure/Application 레이어 분리
  - 의존성 역전 원칙
  - Provider 플러그인 패턴

## 패키지 구조
```
github.com/cagojeiger/cli-recover
├── /cmd           # 실행 파일
│   └── /cli-recover
│       └── /adapters  # Application layer
├── /internal      # 내부 패키지
│   ├── /domain    # 비즈니스 로직
│   ├── /infrastructure  # 외부 시스템
│   └── /application     # 애플리케이션 서비스
├── /testdata      # 테스트 데이터
└── /backup        # 레거시 백업
```

## 메타데이터 저장
- **파일 기반**: ~/.cli-recover/metadata/
- **JSON 형식**: 직렬화/역직렬화
- **SHA256 체크섬**: 무결성 검증

## 진행률 표시
- **채널 기반**: StreamProgress() <-chan Progress
- **실시간 업데이트**: 500ms 간격
- **ETA 계산**: 처리량 기반

## 빌드 시스템
```makefile
# 현재 Makefile 타겟
test         # 전체 테스트
test-coverage # 커버리지 (TUI 제외)
build        # 바이너리 빌드
lint         # golangci-lint
```

## 테스트 전략
- **단위 테스트**: Provider, Registry, Store
- **통합 테스트**: CLI 명령 엔드투엔드
- **Mock 사용**: Kubernetes client, Command executor
- **커버리지 목표**: 80% (TUI 제외됨)

## 제거된 기술 (2025-01-07)
- ~~Bubble Tea~~: TUI 프레임워크 
- ~~Lipgloss~~: 터미널 스타일링
- ~~termenv~~: 터미널 환경
- ~~Bubbles~~: UI 컴포넌트 라이브러리

제거 이유: 헥사고날 아키텍처 위반, God Object 안티패턴, 테스트 불가능한 구조
백업 위치: backup/legacy-tui-20250107/