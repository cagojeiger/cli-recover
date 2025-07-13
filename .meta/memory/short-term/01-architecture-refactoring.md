# 아키텍처 리팩토링 작업

## 작업 일자
2025-07-13

## 작업 배경
- 사용자 피드백: "너무 복잡한 것 같아"
- CLAUDE.md Occam's Razor 원칙 위반 (복잡도 85/100)
- 빈 디렉토리와 미사용 인터페이스 다수

## 수행한 작업

### 1. 디렉토리 구조 단순화
```
이전:
internal/
├── application/
│   ├── port/        (비어있음)
│   └── usecase/
│       └── strategy/
├── domain/
│   ├── entity/
│   ├── repository/  (비어있음)
│   ├── service/
│   └── valueobject/
└── infrastructure/
    ├── adapter/     (비어있음)
    ├── cli/         (비어있음)
    └── persistence/

이후:
internal/
├── pipeline/
│   ├── pipeline.go
│   ├── parser.go
│   ├── builder.go
│   └── executor.go
└── logger/
    └── logger.go
```

### 2. 실행 전략 단순화
- 모든 전략 패턴 제거
- Unix pipe만 사용
- io.Pipe 완전 제거

### 3. 모듈 이름 수정
- github.com/cagojeiger/cli-recover → cli-pipe

### 4. 테스트 작성
- 모든 핵심 기능에 대한 테스트
- 커버리지: 94.3%

## 결과

### 성공
- 모든 examples/*.yaml 정상 작동
- 데드락 없음
- 빌드/테스트 시간 단축

### 메트릭
- 코드 라인: 3,143 → 1,364 (57% 감소)
- 파일 수: 39개 변경
- 복잡도: 85/100 → 35/100

## 다음 단계
- Phase 2: 파라미터 시스템 구현
- Phase 3: TUI 개발