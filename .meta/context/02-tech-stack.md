# cli-pipe 기술 스택

## 핵심 언어
- **Go 1.21+**
  - 이유: 단일 바이너리 배포, 크로스 플랫폼, 우수한 동시성
  - 주요 기능: context, generics, embedded files

## 필수 Unix 커맨드
### 파일/데이터 처리
- **cp**: 로컬 복사, 원자적 작업
- **tar**: 아카이빙, 스트림 생성
- **tee**: 스트림 분기 (핵심 기능)
- **cat**: 파일 읽기/쓰기
- **mv**: 원자적 이동 (임시파일 패턴)

### 스트림 처리
- **gzip/gunzip**: 압축/해제
- **sha256sum**: 체크섬 계산
- **pv**: 진행률 표시
- **head/tail**: 부분 읽기

### 원격/네트워크
- **ssh**: 원격 명령 실행
- **scp**: 원격 파일 복사
- **nc**: 포트 체크

### 프로세스 제어
- **sh/bash**: 복잡한 명령 실행
- **timeout**: 시간 제한
- **kill**: 프로세스 정리
- **mkfifo**: Named pipe

### 검색/필터
- **grep**: 필터링
- **find**: 파일 검색
- **test**: 조건 검사
- **du**: 크기 계산

## Go 라이브러리 (최소화 원칙)
### 표준 라이브러리만 사용
- **os/exec**: 명령 실행
- **io**: 스트림 처리 (io.Pipe 포함)
- **encoding/json**: 설정/로그
- **sync**: 동시성 제어 (WaitGroup, Mutex)
- **flag**: CLI 옵션 파싱

### 외부 라이브러리 (필수)
- **yaml.v3**: YAML 파이프라인 파싱
- **testify**: 테스트 assertion

### 실행 전략 구현
- **ShellPipeStrategy**: Unix pipe 활용 (데드락 방지)
- **GoStreamStrategy**: io.Pipe 기반 (세밀한 제어)
- **전략 패턴**: 자동 선택 또는 수동 지정

## 파일 형식
- **YAML**: 파이프라인 정의
- **JSON**: 로그, 메타데이터
- **SQLite**: 인덱스, 검색

## 개발 도구
- **Make**: 빌드 자동화
- **golangci-lint**: 코드 품질
- **go test**: 단위 테스트
- **testcontainers**: 통합 테스트

## 런타임 의존성
### 필수
- Unix/Linux 환경
- 기본 Unix 도구들

### 선택적 (있으면 사용)
- **kubectl**: Kubernetes 작업
- **docker**: 컨테이너 작업
- **mc**: MinIO 작업

## 버전 관리
- **Semantic Versioning**: v0.1.0 형식
- **Go Modules**: 의존성 관리
- **Git Tags**: 릴리즈 관리

## 성능 목표
- 바이너리 크기: < 20MB
- 메모리 사용: < 100MB (일반 작업)
- 시작 시간: < 100ms
- 스트림 처리: 메모리 상수 사용

## 보안 고려사항
- 민감 정보 마스킹
- 안전한 임시 파일 생성
- 명령 주입 방지
- 파일 권한 보존