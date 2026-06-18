# AI 활용 내역

## 사용 도구

- OpenAI Codex

## 활용 범위

- 과제 요구사항을 애플리케이션, Docker, Kubernetes 영역으로 분해
- `/proc/stat` CPU Delta 계산 방식과 Linux 필드 의미 검토
- 구현 순서 및 시간 배분 정리
- 테스트 케이스와 예외 상황 도출
- Dockerfile과 DaemonSet manifest 검토
- Minikube 배포 오류의 원인 분석과 검증 명령 정리
- README 및 검증 문서 작성 지원

## 구현 과정

1. AI와 요구사항을 분석해 계산식과 Kubernetes volume 전략을 확정했다.
2. 테스트를 먼저 작성하고 실패를 확인한 뒤 Go 구현을 추가했다.
3. macOS에 `/proc`가 없는 문제는 코드로 숨기지 않고 Linux 컨테이너와 Minikube에서 검증했다.
4. Docker image, DaemonSet rollout, Pod 설정과 실제 로그를 직접 실행해 확인했다.
5. 최종 코드와 문서를 과제 원문에 다시 대조했다.
6. 클린 코드 리뷰 후 실행 루프를 테스트 가능한 함수로 분리하고 초기화 실패의 exit code를 보완했다.

## 책임 범위

AI의 제안은 참고 자료로 사용했으며 다음 사항은 지원자가 직접 판단하고 검증했다.

- CPU 계산식과 edge case 처리
- 외부 라이브러리를 사용하지 않는 구현
- Host `/proc` mount와 Downward API 구성
- 테스트, 빌드, 배포 명령의 실제 성공 여부
- 최종 제출 코드와 문서의 적합성

AI가 생성한 설명만으로 완료를 판단하지 않고, 단위 테스트와 실제 Minikube 로그를 완료 기준으로 사용했다.
