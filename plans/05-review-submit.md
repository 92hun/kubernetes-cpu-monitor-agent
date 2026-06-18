# Plan 05 — Review, README, and Submission

**Goal:** 요구사항 누락을 제거하고 재현 가능한 README와 제출 패키지를 만든다.

**Skills:** `requesting-code-review`, `receiving-code-review`, `verification-before-completion`; 저장소 제출 시 `finishing-a-development-branch`

**Time budget:** 10분

## Task 1 — Requirements Review

Critical/Important만 제출 전에 반드시 해결한다.

- [ ] `/proc/stat` 직접 파싱
- [ ] guest 계열 중복 없음
- [ ] 0 Delta와 누적값 감소 처리
- [ ] 첫 로그는 두 유효 샘플 사용
- [ ] `PROC_ROOT` 기본값 `/proc`
- [ ] 5초 주기와 stdout 형식
- [ ] selector와 Pod label 일치
- [ ] hostPath, readOnly, `PROC_ROOT` 연결
- [ ] `NODE_NAME` Downward API
- [ ] 외부 dependency와 privileged 없음

리뷰 의견은 바로 반영하지 않고 과제 원문, 코드, 테스트에 대조한다.

## Task 2 — README

포함 내용:

1. 프로젝트 개요
2. CPU 계산식
3. 파일 구조
4. 사전 요구사항
5. 테스트
6. 이미지 빌드
7. Kubernetes 배포
8. 로그 확인
9. 실제 검증 결과
10. 설계 결정과 Minikube 노드 CPU라는 제한
11. AI 활용 내역

AI 표기:

```markdown
## AI 활용

- 사용 도구: OpenAI Codex
- 활용 범위: 요구사항 분석, 구현 순서, CPU 계산 검토, 테스트·예외 상황, Dockerfile·Manifest 검토
- 최종 코드와 실행 결과는 지원자가 직접 검토하고 Minikube 환경에서 검증했습니다.
```

## Task 3 — Fresh Final Verification

```bash
go test ./...
go vet ./...
go build ./...
docker build -t cpu-monitor:1.0.0 .
kubectl apply --dry-run=client -f agent.yaml
kubectl rollout status daemonset/cpu-monitoring -n whatap --timeout=60s
kubectl get daemonset,pod -n whatap -o wide
kubectl logs -l app=cpu-monitoring -n whatap --tail=5
```

## Task 4 — ZIP

```bash
cd /Users/gim-yeonhun
zip -r cpu-monitor-agent.zip kubernetes-cpu-monitor-agent \
  -x '*/.git/*' '*/.DS_Store' '*/cpu-monitor'
unzip -l cpu-monitor-agent.zip
```

제출 대상에 분석 계획 문서를 포함할지는 최종 판단한다. 필수 파일은 source, `go.mod`, `main_test.go`, `Dockerfile`, `agent.yaml`, `README.md`다.

## PASS Criteria

- 모든 최신 검증 명령 성공
- README만으로 재현 가능
- AI 활용 내역 포함
- ZIP 목록에 필수 파일이 있고 바이너리·임시 파일이 없음

## Result

```text
Status: NOT_STARTED
Evidence:
Changed files:
Next: 이메일 제출
```

