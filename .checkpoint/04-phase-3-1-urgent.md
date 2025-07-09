# Phase 3-1: Restore 긴급 수정

## 날짜: 2025-01-09
## 상태: 🚨 긴급 작업 진행 중
## 브랜치: feature/tui-backup

## 문제 상황

### 사용자 보고
```bash
$ ./cli-recover restore filesystem code-server-67cb5dccf-t8ck7 backup-test2.tar --namespace vscode -t /tmp/test
2025-07-09 13:48:48.074 [INFO ] Analyzing backup file...
2025-07-09 13:48:48.074 [INFO ] Estimated size size=276.1 MB
2025-07-09 13:48:48.074 [INFO ] Starting restore provider=filesystem pod=code-server-67cb5dccf-t8ck7 target_path=/tmp/test
Restore: Starting...
# 여기서 무한 대기 상태
```

### 근본 원인 분석
1. **바이너리 스트리밍 실패**
   - `executor.Stream()` 사용 (텍스트 전용)
   - tar 바이너리 데이터 손상
   - stdin 연결 실패

2. **진행률 표시 없음**
   - tar -x에 verbose 옵션 누락
   - stderr 모니터링 실패
   - 3초 지연으로 인한 무반응

3. **문서 스펙 미준수**
   - StreamBinary 미사용
   - 타임아웃 미적용
   - 구조화된 에러 부재

## 해결 방안

### 1단계: RestoreExecutor 구현
- 새로운 바이너리 안전 executor
- 파일을 직접 stdin으로 연결
- stderr 실시간 모니터링

### 2단계: 진행률 개선
- tar -v 옵션 추가
- 즉각적 피드백 (3초 지연 제거)
- 실시간 파일 복원 상태 표시

### 3단계: 에러 처리
- context.WithTimeout 적용
- 구조화된 에러 메시지
- 파일 경로 검증

## 현재 진행 상황

### 완료된 작업
- [x] 문제 분석 완료
- [x] 해결 방안 수립
- [x] Phase 3-1 계획 문서 작성
- [x] RestoreExecutor 구현 (바이너리 안전)
- [x] restore.go 수정 (새 executor 사용)
- [x] 진행률 표시 개선 (즉시 피드백)
- [x] 에러 처리 강화 (타임아웃, 구조화된 에러)
- [x] 테스트 작성 완료

### 진행 중
- [ ] 문서 업데이트
- [ ] 빌드 및 테스트
- [ ] 커밋 및 푸시

### 대기 중
- [ ] 사용자 검증
- [ ] Phase 3 최종 마무리

## 위험 요소
1. **호환성**: 기존 동작 변경 최소화
2. **성능**: 스트리밍 효율성 유지
3. **안정성**: 네트워크 오류 처리

## 예상 완료 시간
- 목표: 2025-01-09 저녁
- 복잡도: 35/100
- 예상 시간: 4-6시간

## 테스트 계획
1. 단위 테스트
   - RestoreExecutor 테스트
   - 진행률 파싱 테스트
   - 타임아웃 테스트

2. 통합 테스트
   - 실제 pod 복원 테스트
   - 대용량 파일 테스트
   - 네트워크 오류 시뮬레이션

## 성공 기준
- ✅ restore 명령어 즉시 반응
- ✅ 실시간 진행률 표시
- ✅ 바이너리 파일 무손실 복원
- ✅ 적절한 타임아웃 처리

## 다음 단계
Phase 3-1 완료 후:
1. Phase 3 최종 마무리
2. Phase 4 (TUI) 또는 Phase 5 (테스트 커버리지) 선택