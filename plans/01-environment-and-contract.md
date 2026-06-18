# Plan 01 — Environment and Contract

**Goal:** 구현 전에 Docker, Minikube, kubectl을 실행 가능한 상태로 만들고 기술 계약을 고정한다.

**Skills:** `harness-engineering`, `writing-plans`, `verification-before-completion`

**Time budget:** 6분

## Read First

- `../AGENTS.md`
- 이 파일만 읽는다. 이후 계획 파일은 PASS 후 읽는다.

## Technical Contract

CPU 계산:

```text
Idle = idle + iowait
NonIdle = user + nice + system + irq + softirq + steal
Total = Idle + NonIdle
Usage = (TotalDelta - IdleDelta) / TotalDelta * 100
```

`guest`와 `guest_nice`는 이미 `user`와 `nice`에 포함되므로 다시 더하지 않는다.

예정 인터페이스:

```go
type CPUStat struct {
    Idle  uint64
    Total uint64
}

func readCPUStat(procRoot string) (CPUStat, error)
func calculateCPUUsage(previous, current CPUStat) (float64, error)
```

예정 파일:

```text
main.go
main_test.go
go.mod
Dockerfile
agent.yaml
README.md
```

## Execute

- [x] Docker daemon 확인

```bash
docker version
```

- [x] Minikube 시작

```bash
minikube start --driver=docker
```

- [x] 클러스터 상태 확인

```bash
minikube status
kubectl get nodes
```

## PASS Criteria

- Docker client와 server 정보가 모두 출력된다.
- Minikube host, kubelet, apiserver가 Running이다.
- `kubectl get nodes`에서 minikube가 Ready다.
- 위 기술 계약에 TBD나 모호한 선택이 없다.

## Failure Route

Minikube 시작이 실패하면 오류 전체를 보존한다. 같은 명령으로 재현한 뒤 Docker daemon → driver → network → Kubernetes bootstrap 순서로 실패 경계를 찾는다. 임의 설정 변경을 연속으로 시도하지 않는다.

## Handoff

PASS 후 `02-go-core-tdd.md`만 읽는다.

## Result

```text
Status: PASS
Evidence: Docker Desktop server 27.4.0; Minikube host/kubelet/apiserver Running; node minikube Ready on Kubernetes v1.35.1
Changed files: ANALYSIS_AND_PLAN.md, AGENTS.md, plans/01-05
Next: Plan 02 실행
```
