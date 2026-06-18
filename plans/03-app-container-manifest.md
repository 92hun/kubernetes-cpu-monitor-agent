# Plan 03 — Application, Container, and Manifest

**Goal:** 로컬 실행을 확인하고 Dockerfile과 DaemonSet manifest를 완성한다.

**Skills:** `verification-before-completion`; 코드 수정 시 `test-driven-development`; 오류 시 `systematic-debugging`

**Time budget:** 16분

## Inputs

- Plan 02 PASS
- `main.go`, `main_test.go`, `go.mod` 존재

## Task 1 — Local Runtime

macOS에는 `/proc`가 없으므로 `go run`은 파서의 오류 처리를 확인하는 용도로만 사용한다. 5초 루프는 Task 2의 Linux 컨테이너에서 검증한다.

```bash
NODE_NAME=local PROC_ROOT=/proc go run .
```

- [x] 약 5초 뒤 첫 로그
- [x] 이후 약 5초 주기
- [x] `[Host: docker-test] CPU: 12.5%` 형식
- [x] 값이 0.0~100.0
- [x] Ctrl+C 정상 종료

## Task 2 — Dockerfile

요구사항:

- 멀티 스테이지
- builder에서 `CGO_ENABLED=0 GOOS=linux` 빌드
- runtime에는 바이너리만 포함
- 가능하면 non-root
- entrypoint는 CPU monitor 바이너리

검증:

```bash
docker build -t cpu-monitor:1.0.0 .
docker image inspect cpu-monitor:1.0.0
```

## Task 3 — agent.yaml

필수 구조:

- `apiVersion: apps/v1`
- `kind: DaemonSet`
- `metadata.namespace: whatap`
- selector와 Pod label: `app: cpu-monitoring`
- image: `cpu-monitor:1.0.0`
- `imagePullPolicy: Never`
- hostPath `/proc`, type `Directory`
- mountPath `/host/proc`, readOnly
- `PROC_ROOT=/host/proc`
- `NODE_NAME` from `spec.nodeName`
- privileged 미사용

검증:

```bash
kubectl apply --dry-run=client -f agent.yaml
```

## Quality Gate

```bash
go test ./...
go vet ./...
docker build -t cpu-monitor:1.0.0 .
kubectl apply --dry-run=client -f agent.yaml
```

## PASS Criteria

- 로컬 로그가 두 번 이상 확인된다.
- Docker image build 성공.
- YAML client dry-run 성공.
- 이미지명, label, env, volume이 과제 원문과 일치한다.

## Handoff

PASS 후 `04-minikube-deploy-debug.md`만 읽는다.

## Result

```text
Status: PASS
Evidence: Linux container emitted 4.8%, 4.2%, 9.2%, 5.3% at 5-second intervals and exited on Ctrl+C; image sha256:7030ca... is linux/arm64; client dry-run succeeded; go test and go vet exit 0
Changed files: Dockerfile, agent.yaml, plans/03-app-container-manifest.md
Next: Plan 04 실행
```
