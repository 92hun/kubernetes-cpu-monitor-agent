# Plan 02 — Go Core TDD

**Goal:** `/proc/stat` 파서, Delta 계산, 환경 변수와 5초 실행 루프를 TDD로 구현한다.

**Skills:** `test-driven-development`, `verification-before-completion`; 실패 시 `systematic-debugging`

**Time budget:** 24분

## Inputs

- Plan 01 PASS
- Go 표준 라이브러리만 사용
- 함수 시그니처는 Plan 01 계약과 동일

## Task 1 — Project Skeleton

- [x] `go mod init cpu-monitor`
- [x] 빈 `main.go`와 테스트 가능한 패키지 구성
- [x] `go test ./...` 실행 가능 확인

## Task 2 — Parser RED/GREEN

테스트를 먼저 작성한다.

- [x] 정상 stat 샘플에서 Idle과 Total 계산
- [x] 경로가 `<PROC_ROOT>/stat`으로 결정됨
- [x] 필드 부족 오류
- [x] 숫자 파싱 오류
- [x] RED가 기능 부재 때문에 실패하는지 확인

```bash
go test ./... -run TestReadCPUStat -v
```

최소 구현:

- `filepath.Join(procRoot, "stat")`
- 첫 번째 `cpu` 행만 사용
- `strings.Fields`로 공백 처리
- 최소 user~steal 필드 검사
- guest 계열 제외

GREEN:

```bash
go test ./... -run TestReadCPUStat -v
```

## Task 3 — Delta RED/GREEN

테스트를 먼저 작성한다.

- [x] 알려진 previous/current 값의 예상 사용률
- [x] `TotalDelta == 0` 오류
- [x] current 누적값 감소 오류
- [x] 결과 범위 확인

```bash
go test ./... -run TestCalculateCPUUsage -v
```

최소 구현은 underflow 검사 후 Delta를 계산하고 float64 퍼센트를 반환한다.

## Task 4 — Runtime Loop

- [x] `PROC_ROOT` 미설정 시 `/proc`
- [x] `NODE_NAME` 미설정 시 `unknown`
- [x] 최초 값은 기준값으로만 저장
- [x] `time.NewTicker(5 * time.Second)` 사용
- [x] 소수점 한 자리 stdout 로그
- [x] SIGINT와 SIGTERM 종료
- [x] 실행 루프의 sample 오류 복구와 stdout 출력 테스트
- [x] 초기 sample 실패 시 non-zero 종료
- [x] 고정 의존성을 `cpuMonitor`에 묶고 1회 측정과 반복 제어 분리

코드 변경 전 테스트 가능한 로직의 실패 테스트를 먼저 작성한다.

## Quality Gate

```bash
gofmt -w main.go main_test.go
go test ./...
go vet ./...
go build ./...
```

## PASS Criteria

- 모든 테스트가 RED를 거쳐 GREEN이 됐다.
- test, vet, build가 exit code 0이다.
- 외부 dependency가 없다.
- 함수명과 타입이 Plan 01 계약과 일치한다.

## Failure Route

오류를 재현하고 원인 가설 하나만 세운다. 코드 수정 전 실패 테스트를 추가한다. 서로 다른 수정 3회 실패 시 구현 구조를 재검토한다.

## Handoff

PASS 후 `03-app-container-manifest.md`만 읽는다.

## Result

```text
Status: PASS
Evidence: Parser, Delta, env, format, monitor loop, error recovery, and exit-code tests observed RED then GREEN; go test, go vet, go build exit 0; go list -m all shows only cpu-monitor
Changed files: go.mod, main.go, main_test.go
Next: Plan 03 실행
```
