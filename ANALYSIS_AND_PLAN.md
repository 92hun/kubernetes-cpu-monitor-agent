# Kubernetes CPU Monitoring Agent — Execution Index

이 파일은 짧은 라우팅 문서다. 아래 계획 파일을 반드시 번호 순서대로 하나씩 읽고 실행한다. 현재 파일이 PASS가 되기 전에는 다음 파일을 읽거나 실행하지 않는다.

## Global Goal

Go 표준 라이브러리로 `/proc/stat` Delta 기반 CPU 사용률을 계산하고, `cpu-monitor:1.0.0` 이미지와 Kubernetes DaemonSet으로 Minikube에 배포하여 실제 로그를 검증한다.

## Hard Constraints

- `PROC_ROOT` 기본값: `/proc`
- 주기: 5초
- 로그: `[Host: <NODE_NAME>] CPU: <USAGE>%`
- Namespace: `whatap`
- Label: `app=cpu-monitoring`
- Host `/proc` → container `/host/proc`, read-only
- `NODE_NAME` ← Downward API `spec.nodeName`
- 외부 Go 라이브러리와 `privileged: true` 사용 금지
- 최신 실행 증거 없이 PASS 기록 금지

## Sequential Plans

- [x] [01 — 환경과 설계 고정](plans/01-environment-and-contract.md)
- [x] [02 — Go 핵심 로직 TDD](plans/02-go-core-tdd.md)
- [ ] [03 — 애플리케이션·Docker·Manifest](plans/03-app-container-manifest.md)
- [ ] [04 — Minikube 배포·디버깅](plans/04-minikube-deploy-debug.md)
- [ ] [05 — 리뷰·README·제출](plans/05-review-submit.md)

## Progress Record

각 파일 마지막에 다음 형식으로 결과를 남긴다.

```text
Status: PASS | FAIL | BLOCKED
Evidence:
Changed files:
Next:
```

현재 상태: Plan 01~02 PASS. Plan 03 시작 전.
