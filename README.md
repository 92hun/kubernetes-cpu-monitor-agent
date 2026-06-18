# CPU Monitoring Agent

Linux 노드의 `/proc/stat`을 직접 파싱해 5초 간격으로 CPU 사용률을 계산하고, Kubernetes DaemonSet으로 모든 노드에 배포하는 에이전트입니다.

## CPU 계산 방식

`/proc/stat`의 aggregate `cpu` 행에서 다음 누적값을 사용합니다.

```text
Idle = idle + iowait
NonIdle = user + nice + system + irq + softirq + steal
Total = Idle + NonIdle

CPUUsage = (TotalDelta - IdleDelta) / TotalDelta * 100
```

`guest`와 `guest_nice`는 이미 `user`와 `nice`에 포함되므로 중복 합산하지 않습니다.

## 파일 구조

```text
main.go       애플리케이션과 /proc/stat 파서
main_test.go  파서, Delta 계산, 환경 변수, 로그 형식 테스트
go.mod        Go 모듈 정의
Dockerfile    정적 바이너리 멀티 스테이지 빌드
.dockerignore Docker build context 제외 목록
agent.yaml    Kubernetes DaemonSet
docs/         요구사항 분석, 구현 과정, 검증 및 AI 활용 기록
```

## 사전 요구사항

- Go 1.22 이상
- Docker
- kubectl
- Minikube

```bash
minikube start --driver=docker
minikube status
kubectl get nodes
```

## 테스트

```bash
go test ./...
go vet ./...
go build ./...
```

macOS에는 `/proc` 파일 시스템이 없으므로 애플리케이션의 직접 실행은 Linux 환경에서 검증해야 합니다.

## 이미지 빌드

과제에서 제시한 Minikube Docker daemon에 이미지를 빌드합니다.

```bash
eval $(minikube docker-env)
docker build -t cpu-monitor:1.0.0 .
docker image inspect cpu-monitor:1.0.0
eval $(minikube docker-env -u)
```

## 배포

```bash
kubectl create namespace whatap
kubectl apply -f agent.yaml
kubectl rollout status daemonset/cpu-monitoring -n whatap
kubectl get daemonset,pod -n whatap -o wide
```

`whatap` namespace가 이미 존재하면 namespace 생성 명령은 생략합니다.

## 로그 확인

```bash
kubectl logs -l app=cpu-monitoring -n whatap -f
```

출력 형식:

```text
[Host: minikube] CPU: 4.7%
[Host: minikube] CPU: 5.7%
```

## Kubernetes 설계

- DaemonSet으로 노드마다 Pod 하나를 실행합니다.
- 노드의 `/proc`를 컨테이너 `/host/proc`에 read-only로 마운트합니다.
- `PROC_ROOT=/host/proc`를 주입합니다.
- `NODE_NAME`은 Downward API의 `spec.nodeName`으로 주입합니다.
- privileged 권한을 사용하지 않고 non-root로 실행합니다.
- `imagePullPolicy: Never`로 Minikube 내부의 로컬 이미지를 사용합니다.
- 최초 `/proc/stat` 읽기에 실패하면 오류를 출력하고 non-zero로 종료합니다.

Minikube에서 측정되는 값은 macOS 물리 호스트가 아니라 Kubernetes 노드 역할을 하는 Minikube Linux 컨테이너의 CPU 사용률입니다.

## 검증 결과

- Go test, vet, build 성공
- Linux/arm64 Docker 이미지 빌드 성공
- DaemonSet desired/current/ready: `1/1/1`
- Pod 상태: `Running`, restart `0`
- Pod node: `minikube`
- Host `/proc` → `/host/proc`, read-only mount 확인
- 실제 CPU 로그가 5초 간격으로 출력되는 것을 확인

## AI 활용

- 사용 도구: OpenAI Codex
- 활용 범위: 요구사항 분석, 구현 순서 정리, `/proc/stat` CPU 계산 검토, 테스트 및 예외 상황 점검, Dockerfile과 Kubernetes manifest 검토
- 최종 코드와 실행 결과는 지원자가 직접 검토하고 Minikube 환경에서 검증했습니다.

상세한 개발 과정은 다음 문서에 정리했습니다.

- [요구사항 분석](docs/REQUIREMENTS_ANALYSIS.md)
- [구현 계획과 판단](docs/IMPLEMENTATION_PLAN.md)
- [검증 결과](docs/VERIFICATION.md)
- [AI 활용 내역](docs/AI_USAGE.md)
- [AI 단계별 실행 기록](docs/ai-workflow/README.md)
