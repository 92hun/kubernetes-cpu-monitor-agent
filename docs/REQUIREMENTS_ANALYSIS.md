# 요구사항 분석

## 목표

Linux 노드의 `/proc/stat`을 직접 읽어 CPU 사용률을 계산하고, Kubernetes DaemonSet으로 모든 노드에 하나씩 배포한다. 결과는 `kubectl logs`로 즉시 확인할 수 있어야 한다.

## 핵심 요구사항

| 영역 | 요구사항 | 구현 결정 |
|---|---|---|
| 데이터 | Host `/proc/stat` 직접 파싱 | `hostPath`로 `/proc`를 `/host/proc`에 read-only mount |
| 경로 | 하드코딩 금지 | `PROC_ROOT`, 기본값 `/proc` |
| 계산 | 5초 간격 Delta | 최초 값을 기준값으로 저장 후 ticker마다 계산 |
| 노드명 | 실제 노드 이름 | Downward API `spec.nodeName` |
| 출력 | stdout 지정 형식 | `[Host: <NODE_NAME>] CPU: <USAGE>%` |
| 배포 | 모든 노드에 하나 | `apps/v1` DaemonSet |
| 이미지 | `cpu-monitor:1.0.0` | Minikube Docker daemon에 직접 빌드 |

## CPU 계산

`/proc/stat`의 aggregate `cpu` 행은 부팅 이후 누적 CPU 시간을 제공한다.

```text
cpu user nice system idle iowait irq softirq steal guest guest_nice
```

```text
Idle = idle + iowait
NonIdle = user + nice + system + irq + softirq + steal
Total = Idle + NonIdle

TotalDelta = CurrentTotal - PreviousTotal
IdleDelta = CurrentIdle - PreviousIdle
CPUUsage = (TotalDelta - IdleDelta) / TotalDelta * 100
```

`guest`, `guest_nice`는 `user`, `nice`에 이미 포함되어 있으므로 Total에 다시 더하지 않는다.

## 예외 상황

- `/proc/stat` 파일 접근 실패
- aggregate `cpu` 행의 필드 부족
- 누적값 숫자 변환 실패
- Total Delta가 0인 경우
- 재부팅 등으로 누적값이 감소한 경우
- Idle Delta가 Total Delta보다 큰 비정상 입력

최초 샘플을 읽지 못하면 유효한 기준값이 없으므로 non-zero로 종료한다. 실행 중 발생한 일시적인 읽기 또는 계산 오류는 stderr에 기록하고 다음 측정 주기에 다시 시도한다.

## 범위 제한

- `gopsutil` 등 외부 시스템 라이브러리를 사용하지 않는다.
- 개별 CPU core 사용률은 과제 범위에서 제외하고 aggregate CPU만 계산한다.
- Minikube에서 측정되는 대상은 macOS 물리 호스트가 아니라 Kubernetes 노드 역할을 하는 Linux 환경이다.
