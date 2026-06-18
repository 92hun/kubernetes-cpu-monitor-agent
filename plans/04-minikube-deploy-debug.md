# Plan 04 — Minikube Deploy and Debug

**Goal:** 이미지를 Minikube에 빌드하고 DaemonSet을 실제 배포해 노드 CPU 로그를 확인한다.

**Skills:** `harness-engineering`, `verification-before-completion`; 오류 시 `systematic-debugging`

**Time budget:** 16분

## Inputs

- Plan 03 PASS
- Minikube node Ready
- `cpu-monitor:1.0.0` build 가능
- `agent.yaml` dry-run PASS

## Task 1 — Minikube Image

```bash
eval $(minikube docker-env)
docker build -t cpu-monitor:1.0.0 .
docker image inspect cpu-monitor:1.0.0
eval $(minikube docker-env -u)
```

`docker image inspect`는 Minikube Docker 환경이 활성화된 동안 실행한다.

## Task 2 — Deploy

```bash
kubectl get namespace whatap
kubectl create namespace whatap
kubectl apply -f agent.yaml
kubectl rollout status daemonset/cpu-monitoring -n whatap --timeout=60s
kubectl get daemonset,pod -n whatap -o wide
```

namespace가 없을 때만 create한다.

## Task 3 — Runtime Evidence

```bash
kubectl logs -l app=cpu-monitoring -n whatap --tail=5
```

확인:

- [ ] desired/current/ready가 노드 수와 일치
- [ ] Pod Running
- [ ] Pod NODE와 로그 Host 일치
- [ ] 5초 간격 로그
- [ ] CPU 0.0~100.0
- [ ] `PROC_ROOT=/host/proc`
- [ ] `/host/proc/stat` 접근 가능

## Debugging Contract

오류 발생 시 바로 수정하지 않는다.

1. 오류 전체 읽기
2. 동일 명령 재현
3. 실패 경계 결정: image store → Pod scheduling → container start → env/volume → parser
4. 가설 하나와 근거 기록
5. 최소 검증
6. 단일 수정
7. 원래 명령과 회귀 검증

증거:

```bash
kubectl get events -n whatap --sort-by=.lastTimestamp
kubectl describe daemonset cpu-monitoring -n whatap
kubectl describe pod -n whatap <POD_NAME>
kubectl logs -n whatap <POD_NAME>
```

## PASS Criteria

실제 DaemonSet rollout과 요구 형식의 CPU 로그가 모두 확인된다.

## Handoff

PASS 후 `05-review-submit.md`만 읽는다.

## Result

```text
Status: NOT_STARTED
Evidence:
Changed files:
Next: Plan 04 실행
```

