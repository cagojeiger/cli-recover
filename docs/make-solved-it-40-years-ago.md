# Make가 40년 전에 이미 해결한 문제들

## 서문

1976년, Stuart Feldman이 Bell Labs에서 Make를 만들었다. 그로부터 거의 50년이 지난 지금, 우리는 여전히 Make가 해결한 문제들을 새로운 언어와 도구로 재발명하고 있다. cli-pipe도 그 중 하나다.

## 1. Make의 천재적 단순함

### 핵심 개념: 타겟과 의존성

```makefile
target: dependencies
	commands
```

이 단순한 문법으로 Make는:
- 의존성 그래프 자동 구축
- 병렬 실행 최적화
- 증분 빌드
- 순환 의존성 감지

를 모두 해결했다.

## 2. DAG(Directed Acyclic Graph) 분석

### Make의 접근법

```makefile
# Make는 이 간단한 정의로 전체 의존성 그래프를 파악한다
all: program data.gz report.pdf

program: main.o utils.o lib.o
	gcc -o program main.o utils.o lib.o

main.o: main.c common.h
	gcc -c main.c

utils.o: utils.c common.h utils.h
	gcc -c utils.c

lib.o: lib.c lib.h
	gcc -c lib.c

data.gz: raw_data.txt program
	./program < raw_data.txt | gzip > data.gz

report.pdf: data.gz analyze.py
	python analyze.py data.gz
	pdflatex report.tex
```

### 자동 의존성 분석

Make는 위 정의만으로:

1. **토폴로지 정렬**: 실행 순서 자동 결정
2. **병렬 가능성 파악**: main.o, utils.o, lib.o는 동시 컴파일 가능
3. **최소 실행**: 변경된 파일만 재빌드

### cli-pipe의 복잡한 구현

```go
// 수백 줄의 Go 코드로 Make가 자동으로 하는 일을 구현
func buildDependencyGraph(steps []Step) map[string]*Node {
    graph := make(map[string]*Node)
    // ... 복잡한 그래프 구성 로직
}

func (p *Pipeline) IsTree() bool {
    // DFS로 사이클 검증
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    // ... 더 많은 복잡한 로직
}
```

## 3. 병렬 실행의 우아함

### Make의 -j 옵션

```bash
# 단 하나의 옵션으로 병렬 실행
make -j4  # 4개 작업 동시 실행
make -j   # CPU 코어 수만큼 자동 병렬화
```

### 실제 동작 예시

```makefile
# 이 Makefile로
app: frontend backend database
	echo "All components ready"

frontend:
	sleep 3 && echo "Frontend built"

backend:
	sleep 2 && echo "Backend built"

database:
	sleep 4 && echo "Database ready"
```

```bash
$ time make -j3
# 출력 (동시 실행)
sleep 3 && echo "Frontend built"
sleep 2 && echo "Backend built"  
sleep 4 && echo "Database ready"
Backend built
Frontend built
Database ready
All components ready

real    0m4.005s  # 4초 (최대 지연 시간)
```

```bash
$ time make
# 순차 실행시
real    0m9.012s  # 9초 (모든 시간의 합)
```

## 4. 증분 빌드와 타임스탬프

### Make의 똑똑한 타임스탬프 검사

```makefile
# Make는 자동으로 파일 수정 시간을 확인
output.txt: input.txt
	expensive-process < input.txt > output.txt
```

- input.txt가 output.txt보다 새로우면 → 재실행
- output.txt가 더 새로우면 → 건너뜀

### cli-pipe는?

매번 전체 파이프라인을 재실행. 증분 실행 개념 없음.

## 5. 패턴 규칙과 자동 변수

### Make의 강력한 패턴 매칭

```makefile
# 모든 .c 파일을 .o로 변환하는 규칙
%.o: %.c
	$(CC) -c $< -o $@

# 모든 .md 파일을 .html로 변환
%.html: %.md
	pandoc $< -o $@

# 자동 변수
# $@ - 타겟 이름
# $< - 첫 번째 의존성
# $^ - 모든 의존성
# $* - 패턴 매치된 부분
```

### cli-pipe는?

각 단계를 일일이 YAML로 정의해야 함.

## 6. 실제 비교: 동일한 작업 구현

### 작업: 데이터 백업 파이프라인

**Make 버전:**

```makefile
.PHONY: backup clean

BACKUP_DIR = backups/$(shell date +%Y%m%d)
DATA_SOURCES = /etc /home/user/documents /var/log

backup: $(BACKUP_DIR)/archive.tar.gz $(BACKUP_DIR)/archive.sha256
	@echo "Backup complete: $(BACKUP_DIR)"

$(BACKUP_DIR)/archive.tar.gz: $(DATA_SOURCES)
	@mkdir -p $(BACKUP_DIR)
	tar czf $@ $^

$(BACKUP_DIR)/archive.sha256: $(BACKUP_DIR)/archive.tar.gz
	sha256sum $< > $@

# 병렬로 여러 백업 생성
multi-backup: backup-etc backup-home backup-logs

backup-%:
	$(MAKE) backup DATA_SOURCES=/$* BACKUP_DIR=backups/$*-$(shell date +%Y%m%d)

clean:
	rm -rf backups/
```

**cli-pipe 버전:**

```yaml
name: backup-pipeline
steps:
  - name: create-archive
    run: tar czf - /etc /home/user/documents /var/log
    output: archive-stream
    
  - name: save-archive
    run: cat > backups/$(date +%Y%m%d)/archive.tar.gz
    input: archive-stream
    
  - name: generate-hash
    run: sha256sum
    input: archive-stream
    output: hash
    
  - name: save-hash
    run: cat > backups/$(date +%Y%m%d)/archive.sha256
    input: hash
```

더 많은 Go 코드가 필요하고, Make만큼 유연하지 않음.

## 7. Make의 숨겨진 강점들

### 조건부 실행

```makefile
ifdef DEBUG
CFLAGS += -g -O0
else
CFLAGS += -O3
endif

# OS별 다른 명령
ifeq ($(OS),Windows_NT)
    RM = del /Q
else
    RM = rm -f
endif
```

### 함수와 매크로

```makefile
# 문자열 처리 함수
sources := foo.c bar.c baz.c
objects := $(sources:.c=.o)  # foo.o bar.o baz.o

# 사용자 정의 함수
reverse = $(2) $(1)
result := $(call reverse,a,b)  # b a
```

### 재귀적 Make

```makefile
SUBDIRS = lib src test docs

all:
	@for dir in $(SUBDIRS); do \
		$(MAKE) -C $$dir; \
	done
```

## 8. Make가 못하는 것 vs cli-pipe가 못하는 것

### Make의 한계:
- 스트리밍 파이프라인 (파일 기반)
- 실시간 진행률 표시
- 웹 기반 UI
- 분산 실행 (기본적으로)

### cli-pipe의 한계:
- 증분 실행
- 조건부 실행
- 패턴 기반 규칙
- 매크로와 함수
- 수십 년의 검증된 안정성
- 거의 모든 Unix 시스템에 기본 설치

## 9. 역사의 교훈

### Make가 생존한 이유

1. **단순하지만 강력함**: 복잡한 문제를 단순한 규칙으로
2. **선언적**: 무엇을 원하는지만 정의
3. **확장 가능**: 쉘 명령어와 완벽한 통합
4. **검증됨**: 수십 년간 수백만 프로젝트에서 사용

### 새로운 빌드 도구들의 운명

- **Ant** (2000): XML 지옥
- **Maven** (2004): 설정 지옥
- **Gradle** (2012): Groovy/Kotlin DSL 복잡성
- **Bazel** (2015): Google 규모가 아니면 과도함
- **cli-pipe** (2024): Make를 Go로 재발명

## 10. Make로 cli-pipe 기능 구현하기

### 1. 선형 파이프라인

```makefile
pipeline: step1 step2 step3
	@echo "Pipeline complete"

step1:
	@echo "Running step 1" | tee step1.log

step2: step1
	@echo "Running step 2" | tee step2.log

step3: step2
	@echo "Running step 3" | tee step3.log
```

### 2. 분기 파이프라인 (Tree)

```makefile
all: compress checksum

data.tar:
	tar cf data.tar /data

compress: data.tar
	gzip -c data.tar > backup.tar.gz

checksum: data.tar
	sha256sum data.tar > backup.sha256

# make -j2로 실행하면 compress와 checksum이 동시 실행
```

### 3. 파라미터화

```makefile
# 파라미터 전달
SOURCE ?= /data
LEVEL ?= 9

backup:
	tar cf - $(SOURCE) | gzip -$(LEVEL) > backup.tar.gz
```

```bash
# 사용
make backup SOURCE=/home LEVEL=6
```

### 4. 로깅과 모니터링

```makefile
LOGDIR = logs/$(shell date +%Y%m%d_%H%M%S)

%.log: %
	@mkdir -p $(LOGDIR)
	@echo "[$$(date)] Starting $*" | tee -a $(LOGDIR)/pipeline.log
	@$(MAKE) $* 2>&1 | tee $(LOGDIR)/$*.log
	@echo "[$$(date)] Completed $*" | tee -a $(LOGDIR)/pipeline.log
```

## 결론

Make는 1976년에 이미:
- DAG 기반 의존성 관리
- 자동 병렬 실행
- 증분 빌드
- 패턴 매칭
- 조건부 실행

을 우아하게 해결했다.

cli-pipe는 2024년에:
- Make의 일부 기능을 Go로 재구현
- 더 많은 코드로 더 적은 기능 제공
- 증분 실행 없음
- 조건부 실행 없음
- 수백 줄의 Go 코드 필요

> "Those who don't understand Unix are condemned to reinvent it, poorly."  
> — Henry Spencer의 격언을 빌려

**Make를 이해하지 못한 자들은 그것을 형편없이 재발명할 운명이다.**

### 진짜 교훈

새로운 도구를 만들기 전에:
1. 기존 도구가 왜 그렇게 설계되었는지 이해하라
2. 정말로 새로운 문제를 해결하는지 자문하라
3. 단순함의 가치를 과소평가하지 마라

Make는 완벽하지 않다. 하지만 거의 50년 동안 살아남은 이유가 있다.