# 구현 계획과 판단

## 구현 순서

1. Minikube와 Docker 실행 환경 확인
2. `/proc/stat` 파서 테스트 작성 및 구현
3. CPU Delta 계산 테스트 작성 및 구현
4. 환경 변수, 5초 ticker, 종료 시그널 구현
5. Docker 이미지 작성
6. DaemonSet manifest 작성
7. Minikube 배포 및 실제 로그 검증
8. 요구사항 리뷰와 제출 문서 정리

## 코드 구조

과제 규모가 작고 책임이 명확하므로 불필요한 계층을 만들지 않고 `main.go` 하나에 실행 로직을 유지했다.

```go
type cpuStat struct {
    Idle  uint64
    Total uint64
}

type cpuMonitor struct {
    read     cpuReader
    procRoot string
    nodeName string
    stdout   io.Writer
    stderr   io.Writer
}

func readCPUStat(procRoot string) (cpuStat, error)
func calculateCPUUsage(previous, current cpuStat) (float64, error)
```

- `cpuMonitor`: 실행 중 변하지 않는 의존성과 설정을 한곳에서 관리
- `readCPUStat`: 파일 읽기와 파싱
- `calculateCPUUsage`: 두 샘플의 Delta 계산
- `envOrDefault`: 환경 변수 기본값 처리
- `formatCPUUsage`: 출력 형식 고정
- `collect`: 한 번의 CPU 읽기, 계산, stdout/stderr 출력
- `monitor`: 종료 신호와 tick에 따른 반복 제어
- `run`: 5초 ticker의 생성과 수명 관리
- `runMain`: 환경 변수, 종료 시그널, process exit code 결정
- `main`: `runMain`의 exit code로 프로세스 종료

Layered/Clean Architecture는 웹 요청, 도메인 서비스, 저장소 같은 경계가 존재할 때 유효하다. 이 과제에 적용하면 파일과 인터페이스만 늘어나므로 단일 파일을 유지하되, 실행 의존성은 `cpuMonitor`에 묶고 각 메서드의 책임을 분리했다.

## 테스트 전략

구현 전에 기대 동작을 테스트로 작성하고 실패를 확인한 뒤 최소 코드를 추가했다.

- 임시 디렉터리의 `stat` fixture로 `PROC_ROOT` 경로 주입 검증
- 정상 CPU 행, 잘못된 첫 행, 누락 파일과 잘못된 필드 검증
- 알려진 누적값으로 75% 계산 검증
- 0 Delta, 감소한 counter, 비정상 Idle Delta 검증
- 환경 변수 기본값과 로그 포맷 검증
- 최초 샘플 실패, 일시적 읽기·계산 오류 복구, tick별 로그 출력 검증
- context 취소와 tick 채널 종료 시 정상 종료 검증
- 잘못된 측정 간격과 초기화 실패의 non-zero 종료 검증

## 컨테이너 설계

- 멀티 스테이지 Docker build
- `CGO_ENABLED=0`, `GOOS=linux` 정적 바이너리
- `scratch` runtime
- numeric non-root 사용자
- `.dockerignore`로 Git, IDE, 문서, 로컬 바이너리를 build context에서 제외

## Kubernetes 설계

- DaemonSet으로 노드마다 Pod 하나 실행
- Host `/proc` read-only mount
- Downward API로 노드명 주입
- privilege escalation 비활성화 및 capability 제거
- 작은 resource requests/limits 지정
- Minikube 로컬 이미지 사용을 위해 `imagePullPolicy: Never`

## 구현 중 판단 변경

macOS에는 `/proc`가 없어 `go run`으로 실제 CPU loop를 검증할 수 없었다. 코드 변경으로 우회하지 않고 Linux Docker 컨테이너와 Minikube 노드에서 실행 검증하도록 전환했다. 이로써 실제 배포 환경과 동일한 Linux `/proc`를 사용했다.
