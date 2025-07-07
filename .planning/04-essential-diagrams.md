# 핵심 다이어그램

## 1. 백업 실행 플로우

```mermaid
sequenceDiagram
    participant User as 사용자
    participant TUI as TUI
    participant BS as BackupService
    participant BT as BackupType
    participant KC as KubeClient
    participant Pod as K8s Pod
    
    User->>TUI: 백업 옵션 선택
    TUI->>BS: CreateBackup(request)
    BS->>BT: ValidateOptions()
    BS->>BT: BuildCommand()
    BS->>KC: ExecInPod(namespace, pod, cmd)
    KC->>Pod: kubectl exec
    Pod-->>KC: 출력 스트림
    KC-->>BS: io.Reader
    BS-->>TUI: Job 상태 업데이트
    TUI-->>User: 진행률 표시
```

## 2. 현재 문제: God Object

```mermaid
graph TD
    Model[Model<br/>115+ fields]
    Model --> UI[UI 상태<br/>- selected<br/>- screen<br/>- width/height]
    Model --> BIZ[비즈니스 로직<br/>- backupOptions<br/>- compressionType<br/>- excludePatterns]
    Model --> DATA[데이터<br/>- namespaces<br/>- pods<br/>- containers]
    Model --> DEPS[의존성<br/>- runner<br/>- jobManager<br/>- kubeClient]
    
    style Model fill:#ff6666
```

## 3. 개선: 계층 분리

```mermaid
graph TB
    subgraph "Presentation Layer"
        TUI[Bubble Tea TUI]
        COMP[재사용 컴포넌트]
    end
    
    subgraph "Domain Layer"
        BS[BackupService]
        JM[JobManager]
        BT[BackupType<br/>플러그인]
    end
    
    subgraph "Infrastructure Layer"
        KC[KubectlAdapter]
        FS[FileSystem]
        CFG[Config]
    end
    
    TUI --> BS
    TUI --> JM
    BS --> BT
    BS --> KC
    JM --> FS
    
    style TUI fill:#90EE90
    style BS fill:#87CEEB
    style KC fill:#FFE4B5
```

## 4. Ring Buffer 메모리 관리

```mermaid
sequenceDiagram
    participant Exec as 백업 프로세스
    participant RB as Ring Buffer<br/>(1000줄)
    participant File as 로그 파일
    participant TUI as TUI Display
    
    loop 백업 진행 중
        Exec->>RB: Write(새 라인)
        alt 버퍼 가득 참
            RB->>RB: 가장 오래된 라인 제거
        end
        RB->>File: Append(전체 로그)
        TUI->>RB: ReadTail(50)
        RB-->>TUI: 최근 50줄
    end
    
    Note over RB: 메모리 사용량 제한
    Note over File: 전체 이력 보존
```

## 5. BackupType 플러그인 확장

```mermaid
graph LR
    subgraph "Core"
        REG[BackupTypeRegistry]
        INT[BackupType<br/>인터페이스]
    end
    
    subgraph "Plugins"
        FS[FilesystemBackup<br/>tar/gzip]
        MIO[MinIOBackup<br/>S3 명령]
        MDB[MongoDBBackup<br/>mongodump]
    end
    
    subgraph "미래 확장"
        PG[PostgreSQLBackup]
        ES[ElasticsearchBackup]
    end
    
    REG --> INT
    FS -.-> INT
    MIO -.-> INT
    MDB -.-> INT
    PG -.-> INT
    ES -.-> INT
    
    style REG fill:#FFE4B5
    style INT fill:#E6E6FA
    style FS fill:#90EE90
    style MIO fill:#90EE90
    style MDB fill:#90EE90
    style PG fill:#D3D3D3
    style ES fill:#D3D3D3
```

## 다이어그램 설명

### 1. **백업 실행 플로우**
- 사용자 선택부터 Pod 내부 실행까지 전체 흐름
- 핵심: kubectl exec를 통한 Pod 접근

### 2. **God Object 문제**
- 현재 Model이 모든 책임을 가짐
- 테스트와 유지보수가 어려움

### 3. **계층 분리 개선**
- 명확한 레이어 구분
- 각 레이어는 인터페이스로만 통신

### 4. **Ring Buffer 메모리 관리**
- 메모리 사용량을 1000줄로 제한
- 전체 로그는 파일로 보존

### 5. **플러그인 확장성**
- 새 백업 타입 추가가 용이
- 기존 코드 수정 없이 확장 가능