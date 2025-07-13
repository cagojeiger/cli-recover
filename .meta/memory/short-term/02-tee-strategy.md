# Hidden tee 전략 발견

## 날짜
2025-07-13

## 발견 과정

### 문제 인식
- YAML의 input/output이 실제로는 무시되고 있음
- 단순히 순서대로 파이프 연결만 하는 상태
- Phase 1의 스트림 분기 기능 구현 필요

### 핵심 통찰
"tee를 YAML에서 숨기고 자동으로 처리하면 어떨까?"

### 해결책
1. 사용자는 단순히 input/output만 정의
2. 내부에서 자동으로 스트림 사용 분석
3. 분기가 필요한 곳에 tee 자동 삽입
4. 복잡한 Unix 명령은 내부에서 처리

## 구현 아이디어

### Stream 분석
```go
// archive가 compress와 checksum 두 곳에서 사용됨 감지
usage["archive"] = &StreamUsage{
    producer: "source",
    consumers: ["compress", "checksum"]
}
```

### 자동 명령어 생성
```bash
# YAML은 단순하게
steps:
  - run: tar cf - /data
    output: archive
  - run: gzip -9
    input: archive
  - run: sha256sum  
    input: archive

# 내부에서 스마트하게
tar cf - /data | tee >(sha256sum > hash.txt) | gzip -9 > backup.gz
```

## 장점

### 사용자 경험
- YAML 문법 변화 없음
- 복잡한 Unix 명령 학습 불필요
- 직관적인 데이터 흐름

### 기술적 장점
- tee는 POSIX 표준 (어디서나 동작)
- 메모리 효율적 (스트리밍)
- 디버깅 가능 (생성된 스크립트 확인)

## 복잡도 평가
- 사용자 관점: 20/100 ✅
- 내부 구현: 40/100 ✅
- 전체: 40/100 (수용 가능)

## 다음 단계
1. StreamAnalyzer 구현
2. SmartBuilder 구현
3. 테스트 케이스 작성
4. 문서화