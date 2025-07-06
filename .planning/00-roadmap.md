# CLI-Restore Roadmap

## Phase 1: MVP (v0.1.0) - Current
- [x] Context Engineering 구조 설정
- [ ] 기본 프로젝트 구조
- [ ] `--version` 명령어 구현
- [ ] 빌드 시스템 (Makefile)
- [ ] GitHub Actions 자동 릴리즈

## Phase 2: Core Features (v0.2.0)
- [ ] 프로젝트 구조 확장 (internal/, pkg/)
- [ ] `backup` 명령어 구현
- [ ] Kubernetes 클라이언트 연동
- [ ] Pod 파일 시스템 접근

## Phase 3: Archive Features (v0.3.0)
- [ ] tar 압축 구현
- [ ] 분할 압축 기능
- [ ] 진행 상황 표시
- [ ] 에러 처리 강화

## Phase 4: Polish (v1.0.0)
- [ ] 설정 파일 지원
- [ ] 다중 Pod 백업
- [ ] 복원 기능
- [ ] Homebrew 배포