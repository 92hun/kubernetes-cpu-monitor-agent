# 검증 결과

검증일: 2026-06-18

## 환경

- macOS arm64
- Minikube v1.38.1
- Kubernetes v1.35.1
- Docker Desktop 4.37.2
- Minikube driver: Docker

## Go 검증

```bash
go test -count=1 ./...
go vet ./...
go build ./...
```

결과:

```text
ok  cpu-monitor
```

세 명령 모두 exit code 0을 확인했다. `go list -m all`에는 `cpu-monitor`만 출력되어 외부 Go dependency가 없음을 확인했다.

Race detector를 포함한 테스트 coverage는 90.3%이며, 핵심 함수별 결과는 다음과 같다.

```text
readCPUStat          100.0%
calculateCPUUsage     88.9%
envOrDefault         100.0%
formatCPUUsage       100.0%
monitor               83.3%
run                  100.0%
runMain               85.7%
```

## Docker 검증

```bash
docker build -t cpu-monitor:1.0.0 .
docker image inspect cpu-monitor:1.0.0
```

결과:

```text
OS: linux
Architecture: arm64
Image: sha256:d4bd0908cea83e5e4689c20691a537b7646f833c4d199576c412d22788b330b4
```

Linux 컨테이너에서 5초 간격 로그와 SIGINT 종료를 확인했다.

## Manifest 검증

```bash
kubectl apply --dry-run=client -f agent.yaml
```

```text
daemonset.apps/cpu-monitoring unchanged (dry run)
```

## Kubernetes 검증

```bash
kubectl rollout status daemonset/cpu-monitoring -n whatap --timeout=60s
kubectl get daemonset,pod -n whatap -o wide
```

결과:

```text
daemon set "cpu-monitoring" successfully rolled out
DaemonSet desired/current/ready: 1/1/1
Pod status: Running
Node: minikube
Restarts: 0
```

Pod spec에서 다음 설정도 확인했다.

```text
PROC_ROOT=/host/proc
mount=/host/proc
readOnly=true
hostPath=/proc
```

## 실제 로그

```bash
kubectl logs -l app=cpu-monitoring -n whatap --tail=5
```

```text
[Host: minikube] CPU: 4.4%
[Host: minikube] CPU: 4.1%
[Host: minikube] CPU: 4.3%
[Host: minikube] CPU: 4.6%
[Host: minikube] CPU: 4.5%
```

노드명, 출력 형식, 값의 범위와 주기적 출력을 확인했다.
